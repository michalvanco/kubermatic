/*
Copyright 2020 The Kubermatic Kubernetes Platform contributors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package mutation

import (
	"context"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Masterminds/semver/v3"
	"github.com/go-logr/logr"

	kubermaticv1 "k8c.io/kubermatic/v2/pkg/crd/kubermatic/v1"
	"k8c.io/kubermatic/v2/pkg/defaulting"
	"k8c.io/kubermatic/v2/pkg/provider"
	"k8c.io/kubermatic/v2/pkg/provider/cloud"

	admissionv1 "k8s.io/api/admission/v1"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"k8s.io/utils/pointer"
	ctrlruntime "sigs.k8s.io/controller-runtime"
	ctrlruntimeclient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// AdmissionHandler for mutating Kubermatic Cluster CRD.
type AdmissionHandler struct {
	log     logr.Logger
	decoder *admission.Decoder

	client       ctrlruntimeclient.Client
	seedGetter   provider.SeedGetter
	configGetter provider.KubermaticConfigurationGetter
	caBundle     *x509.CertPool

	// disableProviderMutation is only for unit tests, to ensure no
	// provide would phone home to validate dummy test credentials
	disableProviderMutation bool
}

// NewAdmissionHandler returns a new cluster mutation AdmissionHandler.
func NewAdmissionHandler(client ctrlruntimeclient.Client, configGetter provider.KubermaticConfigurationGetter, seedGetter provider.SeedGetter, caBundle *x509.CertPool) *AdmissionHandler {
	return &AdmissionHandler{
		client:       client,
		configGetter: configGetter,
		seedGetter:   seedGetter,
		caBundle:     caBundle,
	}
}

func (h *AdmissionHandler) SetupWebhookWithManager(mgr ctrlruntime.Manager) {
	mgr.GetWebhookServer().Register("/mutate-kubermatic-k8s-io-cluster", &webhook.Admission{Handler: h})
}

func (h *AdmissionHandler) InjectLogger(l logr.Logger) error {
	h.log = l.WithName("cluster-mutation-handler")
	return nil
}

func (h *AdmissionHandler) InjectDecoder(d *admission.Decoder) error {
	h.decoder = d
	return nil
}

func (h *AdmissionHandler) Handle(ctx context.Context, req webhook.AdmissionRequest) webhook.AdmissionResponse {
	cluster := &kubermaticv1.Cluster{}
	oldCluster := &kubermaticv1.Cluster{}

	switch req.Operation {
	case admissionv1.Create:
		if err := h.decoder.Decode(req, cluster); err != nil {
			return admission.Errored(http.StatusBadRequest, err)
		}

		err := h.applyDefaults(ctx, cluster)
		if err != nil {
			h.log.Info("cluster mutation failed", "error", err)
			return webhook.Errored(http.StatusInternalServerError, fmt.Errorf("cluster mutation request %s failed: %v", req.UID, err))
		}

		if err := h.mutateCreate(cluster); err != nil {
			h.log.Info("cluster mutation failed", "error", err)
			return webhook.Errored(http.StatusInternalServerError, fmt.Errorf("cluster mutation request %s failed: %v", req.UID, err))
		}

	case admissionv1.Update:
		if err := h.decoder.Decode(req, cluster); err != nil {
			return admission.Errored(http.StatusBadRequest, err)
		}
		if err := h.decoder.DecodeRaw(req.OldObject, oldCluster); err != nil {
			return admission.Errored(http.StatusBadRequest, err)
		}

		// Prevent unconditional CNI upgrades for old clusters.
		// Do not default cni if it was not explicitly set.
		// Todo: Temporary hack will be removed soon
		preventCNIDefaulting := cluster.Spec.CNIPlugin == nil

		// apply defaults to the existing clusters
		err := h.applyDefaults(ctx, cluster)
		if err != nil {
			h.log.Info("cluster mutation failed", "error", err)
			return webhook.Errored(http.StatusInternalServerError, fmt.Errorf("cluster mutation request %s failed: %v", req.UID, err))
		}

		if preventCNIDefaulting {
			cluster.Spec.CNIPlugin = nil
		}

		if err := h.mutateUpdate(oldCluster, cluster); err != nil {
			h.log.Info("cluster mutation failed", "error", err)
			return webhook.Errored(http.StatusInternalServerError, fmt.Errorf("cluster mutation request %s failed: %v", req.UID, err))
		}

	case admissionv1.Delete:
		return webhook.Allowed(fmt.Sprintf("no mutation done for request %s", req.UID))

	default:
		return admission.Errored(http.StatusBadRequest, fmt.Errorf("%s not supported on cluster resources", req.Operation))
	}

	mutatedCluster, err := json.Marshal(cluster)
	if err != nil {
		return webhook.Errored(http.StatusInternalServerError, fmt.Errorf("marshaling cluster object failed: %v", err))
	}

	return admission.PatchResponseFromRaw(req.Object.Raw, mutatedCluster)
}

func (h *AdmissionHandler) applyDefaults(ctx context.Context, c *kubermaticv1.Cluster) error {
	seed, provider, fieldErr := h.buildDefaultingDependencies(ctx, c)
	if fieldErr != nil {
		return fieldErr
	}

	config, err := h.configGetter(ctx)
	if err != nil {
		return err
	}

	defaultTemplate, err := defaulting.GetDefaultingClusterTemplate(ctx, h.client, seed)
	if err != nil {
		return err
	}

	return defaulting.DefaultClusterSpec(&c.Spec, defaultTemplate, seed, config, provider)
}

// mutateCreate is an addition to regular defaulting for new clusters.
// at the time of writing it handles features that should only be enabled for new clusters.
func (h *AdmissionHandler) mutateCreate(newCluster *kubermaticv1.Cluster) error {

	if newCluster.Spec.Features == nil {
		newCluster.Spec.Features = map[string]bool{}
	}

	// Network policies for Apiserver are deployed by default
	if _, ok := newCluster.Spec.Features[kubermaticv1.ApiserverNetworkPolicy]; !ok {
		newCluster.Spec.Features[kubermaticv1.ApiserverNetworkPolicy] = true
	}

	// Network policies for kube-system are deployed by default
	if _, ok := newCluster.Spec.Features[kubermaticv1.KubeSystemNetworkPolicies]; !ok {
		newCluster.Spec.Features[kubermaticv1.KubeSystemNetworkPolicies] = true
	}

	return nil
}

func (h *AdmissionHandler) mutateUpdate(oldCluster, newCluster *kubermaticv1.Cluster) error {
	// This part of the code handles the CCM/CSI migration. It currently works
	// only for OpenStack clusters, in the following way:
	//   * Add the CCM/CSI migration annotations
	//   * Enable the UseOctaiva flag
	if v, oldV := newCluster.Spec.Features[kubermaticv1.ClusterFeatureExternalCloudProvider],
		oldCluster.Spec.Features[kubermaticv1.ClusterFeatureExternalCloudProvider]; v && !oldV {

		switch {
		case newCluster.Spec.Cloud.Openstack != nil:
			addCCMCSIMigrationAnnotations(newCluster)
			newCluster.Spec.Cloud.Openstack.UseOctavia = pointer.BoolPtr(true)

		case newCluster.Spec.Cloud.VSphere != nil:
			addCCMCSIMigrationAnnotations(newCluster)
		}
	}

	// This part handles CNI upgrade from unspecified (= very old) CNI version to the default Canal version.
	// This upgrade is necessary for k8s versions >= 1.22, where v1beta1 CRDs are not supported anymore.
	if newCluster.Spec.CNIPlugin == nil {
		upgradeConstraint, err := semver.NewConstraint(">= 1.22")
		if err != nil {
			return fmt.Errorf("parsing CNI upgrade constraint failed: %v", err)
		}
		if newCluster.Spec.Version.String() != "" && upgradeConstraint.Check(newCluster.Spec.Version.Semver()) {
			newCluster.Spec.CNIPlugin = &kubermaticv1.CNIPluginSettings{
				Type:    kubermaticv1.CNIPluginTypeCanal,
				Version: defaulting.DefaultCNIPluginVersions[kubermaticv1.CNIPluginTypeCanal],
			}
		}
	}

	return nil
}

func addCCMCSIMigrationAnnotations(cluster *kubermaticv1.Cluster) {
	if cluster.ObjectMeta.Annotations == nil {
		cluster.ObjectMeta.Annotations = map[string]string{}
	}

	cluster.ObjectMeta.Annotations[kubermaticv1.CCMMigrationNeededAnnotation] = ""
	cluster.ObjectMeta.Annotations[kubermaticv1.CSIMigrationNeededAnnotation] = ""
}

func (h *AdmissionHandler) buildDefaultingDependencies(ctx context.Context, c *kubermaticv1.Cluster) (*kubermaticv1.Seed, provider.CloudProvider, *field.Error) {
	seed, err := h.seedGetter()
	if err != nil {
		return nil, nil, field.InternalError(nil, err)
	}

	if h.disableProviderMutation {
		return seed, nil, nil
	}

	datacenter, fieldErr := defaulting.DatacenterForClusterSpec(&c.Spec, seed)
	if fieldErr != nil {
		return nil, nil, fieldErr
	}

	secretKeySelectorFunc := provider.SecretKeySelectorValueFuncFactory(ctx, h.client)
	cloudProvider, err := cloud.Provider(datacenter, secretKeySelectorFunc, h.caBundle)
	if err != nil {
		return nil, nil, field.InternalError(nil, err)
	}

	return seed, cloudProvider, nil
}
