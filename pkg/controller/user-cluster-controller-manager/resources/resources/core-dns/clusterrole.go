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

package coredns

import (
	"k8c.io/kubermatic/v2/pkg/resources"
	"k8c.io/kubermatic/v2/pkg/resources/reconciling"

	rbacv1 "k8s.io/api/rbac/v1"
)

// ClusterRoleCreator returns the func to create/update the ClusterRole for CoreDNS.
func ClusterRoleCreator() reconciling.NamedClusterRoleCreatorGetter {
	return func() (string, reconciling.ClusterRoleCreator) {
		return resources.CoreDNSClusterRoleName, func(cr *rbacv1.ClusterRole) (*rbacv1.ClusterRole, error) {
			cr.Rules = []rbacv1.PolicyRule{
				{
					APIGroups: []string{"discovery.k8s.io"},
					Resources: []string{"endpointslices"},
					Verbs:     []string{"list", "watch"},
				},
				{
					APIGroups: []string{""},
					Resources: []string{
						"endpoints",
						"services",
						"pods",
						"namespaces",
					},
					Verbs: []string{"list", "watch"},
				},
			}
			return cr, nil
		}
	}
}
