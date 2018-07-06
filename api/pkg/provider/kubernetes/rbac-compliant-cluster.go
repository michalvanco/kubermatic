package kubernetes

import (
	"errors"
	"strings"
	"time"

	"github.com/golang/glog"

	kubermaticclientset "github.com/kubermatic/kubermatic/api/pkg/crd/client/clientset/versioned"
	kubermaticclientv1 "github.com/kubermatic/kubermatic/api/pkg/crd/client/clientset/versioned/typed/kubermatic/v1"
	kubermaticv1lister "github.com/kubermatic/kubermatic/api/pkg/crd/client/listers/kubermatic/v1"
	kubermaticapiv1 "github.com/kubermatic/kubermatic/api/pkg/crd/kubermatic/v1"
	"github.com/kubermatic/kubermatic/api/pkg/provider"
	machineclientset "github.com/kubermatic/machine-controller/pkg/client/clientset/versioned"

	corev1 "k8s.io/api/core/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/rand"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

// NewRBACCompliantClusterProvider returns a new cluster provider that respects RBAC policies
// it uses createSeedImpersonatedClient to create a connection that uses user impersonation
func NewRBACCompliantClusterProvider(
	createSeedImpersonatedClient kubermaticImpersonationClient,
	seedPrivilegedClient kubermaticclientset.Interface,
	userClusterConnProvider UserClusterConnectionProvider,
	clusterLister kubermaticv1lister.ClusterLister,
	addons []string,
	workerName string) *RBACCompliantClusterProvider {
	return &RBACCompliantClusterProvider{
		createSeedImpersonatedClient: createSeedImpersonatedClient,
		seedPrivilegedClient:         seedPrivilegedClient,
		userClusterConnProvider:      userClusterConnProvider,
		clusterLister:                clusterLister,
		addons:                       addons,
		workerName:                   workerName,
	}
}

// RBACCompliantClusterProvider struct that holds required components in order to provide
// cluster provided that is RBAC compliant
type RBACCompliantClusterProvider struct {
	// createSeedImpersonatedClient is used as a ground for impersonation
	// whenever a connection to Seed API server is required
	createSeedImpersonatedClient kubermaticImpersonationClient

	// userClusterConnProvider used for obtaining a connection to the client's cluster
	userClusterConnProvider UserClusterConnectionProvider

	// clusterLister provide access to local cache that stores cluster objects
	clusterLister kubermaticv1lister.ClusterLister

	// seedPrivilegedClient a privileged client connection used for creating addons only
	seedPrivilegedClient kubermaticclientset.Interface

	addons []string

	workerName string
}

// New creates a brand new cluster that is bound to the given project
func (p *RBACCompliantClusterProvider) New(project *kubermaticapiv1.Project, user *kubermaticapiv1.User, spec *kubermaticapiv1.ClusterSpec) (*kubermaticapiv1.Cluster, error) {
	if project == nil || user == nil || spec == nil {
		return nil, errors.New("project and/or user and/or spec is missing but required")
	}
	spec.HumanReadableName = strings.TrimSpace(spec.HumanReadableName)
	spec.WorkerName = p.workerName

	name := rand.String(10)
	cluster := &kubermaticapiv1.Cluster{
		ObjectMeta: metav1.ObjectMeta{
			Labels: map[string]string{
				kubermaticapiv1.WorkerNameLabelKey: p.workerName,
			},
			Name: name,
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion: kubermaticapiv1.SchemeGroupVersion.String(),
					Kind:       kubermaticapiv1.ProjectKindName,
					UID:        project.GetUID(),
					Name:       project.Name,
				},
			},
		},
		Spec: *spec,
		Status: kubermaticapiv1.ClusterStatus{
			UserEmail:     user.Spec.Email,
			UserName:      user.Name,
			NamespaceName: NamespaceName(name),
		},
		Address: kubermaticapiv1.ClusterAddress{},
	}

	seedImpersonatedClient, err := p.createSeedImpersonationClientWrapper(user, project)
	if err != nil {
		return nil, err
	}
	cluster, err = seedImpersonatedClient.Clusters().Create(cluster)
	if err != nil {
		return nil, err
	}

	//We wait until the cluster exists in the lister so we can use this instead of doing api calls
	existsInLister := func() (bool, error) {
		_, err := p.clusterLister.Get(cluster.Name)
		if err != nil {
			return false, nil
		}
		return true, nil
	}

	// TODO: Make Addons to be part of the cluster specification
	//       For more details see: https://github.com/kubermatic/kubermatic/issues/1211
	// TODO: Add RBAC Roles to `Addons` resources
	//       For more details see: https://github.com/kubermatic/kubermatic/issues/1181
	// TODO: this code deserves refactoring, primarily because we are creating add-ons in a namespace which in fact is created by controller that listens for new clusters.
	go func() {
		gv := kubermaticapiv1.SchemeGroupVersion
		ownerRef := *metav1.NewControllerRef(cluster, gv.WithKind("Cluster"))
		err = wait.Poll(50*time.Millisecond, 60*time.Second, func() (done bool, err error) {
			for _, addon := range p.addons {
				_, err = p.seedPrivilegedClient.KubermaticV1().Addons(cluster.Status.NamespaceName).Create(&kubermaticapiv1.Addon{
					ObjectMeta: metav1.ObjectMeta{
						Name:            addon,
						Namespace:       cluster.Status.NamespaceName,
						OwnerReferences: []metav1.OwnerReference{ownerRef},
					},
					Spec: kubermaticapiv1.AddonSpec{
						Name: addon,
						Cluster: corev1.ObjectReference{
							Name:       cluster.Name,
							Namespace:  "",
							UID:        cluster.UID,
							APIVersion: cluster.APIVersion,
							Kind:       "Cluster",
						},
					},
				})
				if err != nil {
					if kerrors.IsAlreadyExists(err) {
						continue
					}
					glog.V(0).Infof("failed to create initial adddon %s for cluster %s: %v", addon, cluster.Name, err)
					return false, nil
				}
			}

			return true, nil
		})
		if err != nil {
			glog.V(0).Infof("failed to create initial addons in cluster %s: %v", cluster.Name, err)
		}
	}()

	return cluster, wait.Poll(10*time.Millisecond, 30*time.Second, existsInLister)
}

// List gets all clusters that belong to the given project
// If you want to filter the result please take a look at ClusterListOptions
//
// Note:
// After we get the list of clusters we could try to get each cluster individually using unprivileged account to see if the user have read access,
// We don't do this because we assume that if the user was able to get the project (argument) it has to have at least read access.
func (p *RBACCompliantClusterProvider) List(project *kubermaticapiv1.Project, options *provider.ClusterListOptions) ([]*kubermaticapiv1.Cluster, error) {
	if project == nil {
		return nil, errors.New("project is missing but required")
	}
	clusters, err := p.clusterLister.List(labels.Everything())
	if err != nil {
		return nil, err
	}

	projectClusters := []*kubermaticapiv1.Cluster{}
	for _, cluster := range clusters {
		owners := cluster.GetOwnerReferences()
		for _, owner := range owners {
			if owner.APIVersion == kubermaticapiv1.SchemeGroupVersion.String() && owner.Kind == kubermaticapiv1.ProjectKindName && owner.Name == project.Name {
				projectClusters = append(projectClusters, cluster)
			}
		}
	}

	if options == nil {
		return projectClusters, nil
	}
	if len(options.SortBy) > 0 {
		var err error
		projectClusters, err = p.sortBy(projectClusters, options.SortBy)
		if err != nil {
			return nil, err
		}
	}
	if len(options.ClusterName) == 0 {
		return projectClusters, nil
	}

	filteredProjectClusters := []*kubermaticapiv1.Cluster{}
	for _, projectCluster := range projectClusters {
		if projectCluster.Spec.HumanReadableName == options.ClusterName {
			filteredProjectClusters = append(filteredProjectClusters, projectCluster)
		}
	}

	return filteredProjectClusters, nil
}

// Get returns the given cluster, it uses the projectInternalName to determine the group the user belongs to
func (p *RBACCompliantClusterProvider) Get(user *kubermaticapiv1.User, project *kubermaticapiv1.Project, clusterName string) (*kubermaticapiv1.Cluster, error) {
	seedImpersonatedClient, err := p.createSeedImpersonationClientWrapper(user, project)
	if err != nil {
		return nil, err
	}
	return seedImpersonatedClient.Clusters().Get(clusterName, metav1.GetOptions{})
}

// Delete deletes the given cluster
func (p *RBACCompliantClusterProvider) Delete(user *kubermaticapiv1.User, project *kubermaticapiv1.Project, clusterName string) error {
	seedImpersonatedClient, err := p.createSeedImpersonationClientWrapper(user, project)
	if err != nil {
		return err
	}

	return seedImpersonatedClient.Clusters().Delete(clusterName, &metav1.DeleteOptions{})
}

// Update updates a cluster
func (p *RBACCompliantClusterProvider) Update(user *kubermaticapiv1.User, project *kubermaticapiv1.Project, newCluster *kubermaticapiv1.Cluster) (*kubermaticapiv1.Cluster, error) {
	seedImpersonatedClient, err := p.createSeedImpersonationClientWrapper(user, project)
	if err != nil {
		return nil, err
	}

	return seedImpersonatedClient.Clusters().Update(newCluster)
}

// GetAdminKubeconfigForCustomerCluster returns the admin kubeconfig for the given cluster
func (p *RBACCompliantClusterProvider) GetAdminKubeconfigForCustomerCluster(c *kubermaticapiv1.Cluster) (*clientcmdapi.Config, error) {
	b, err := p.userClusterConnProvider.GetAdminKubeconfig(c)
	if err != nil {
		return nil, err
	}

	return clientcmd.Load(b)
}

// GetMachineClientForCustomerCluster returns a client to interact with machine resources in the given cluster
//
// Note that the client you will get has admin privileges
func (p *RBACCompliantClusterProvider) GetMachineClientForCustomerCluster(c *kubermaticapiv1.Cluster) (machineclientset.Interface, error) {
	return p.userClusterConnProvider.GetMachineClient(c)
}

// GetClientForCustomerCluster returns a client to interact with the given cluster
//
// Note that the client you will get has admin privileges
func (p *RBACCompliantClusterProvider) GetClientForCustomerCluster(c *kubermaticapiv1.Cluster) (kubernetes.Interface, error) {
	return p.userClusterConnProvider.GetClient(c)
}

// createSeedImpersonationClientWrapper is a helper method that spits back kubermatic client that uses user impersonation
func (p *RBACCompliantClusterProvider) createSeedImpersonationClientWrapper(user *kubermaticapiv1.User, project *kubermaticapiv1.Project) (kubermaticclientv1.KubermaticV1Interface, error) {
	if user == nil || project == nil {
		return nil, errors.New("user and/or project is missing but required")
	}
	groupName, err := user.GroupForProject(project.Name)
	if err != nil {
		return nil, kerrors.NewForbidden(schema.GroupResource{}, project.Name, err)
	}
	impersonationCfg := restclient.ImpersonationConfig{
		UserName: user.Spec.Email,
		Groups:   []string{groupName},
	}
	return p.createSeedImpersonatedClient(impersonationCfg)
}

// sortBy sort the given clusters by the specified field name (sortBy param)
func (p *RBACCompliantClusterProvider) sortBy(clusters []*kubermaticapiv1.Cluster, sortBy string) ([]*kubermaticapiv1.Cluster, error) {
	rawKeys := []runtime.Object{}
	for index := range clusters {
		rawKeys = append(rawKeys, clusters[index])
	}
	sorter, err := sortObjects(scheme.Codecs.UniversalDecoder(), rawKeys, sortBy)
	if err != nil {
		return nil, err
	}

	sortedClusters := make([]*kubermaticapiv1.Cluster, len(clusters))
	for index := range clusters {
		sortedClusters[index] = clusters[sorter.originalPosition(index)]
	}
	return sortedClusters, nil
}
