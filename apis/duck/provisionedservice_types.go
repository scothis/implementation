/*
 * Copyright 2023 Original Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      https://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package duck

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ProvisionedServiceStatus defines the observed state of ServiceBinding
type ProvisionedServiceStatus struct {
	// Binding exposes the projected secret for this ServiceBinding
	Binding corev1.LocalObjectReference `json:"binding,omitempty"`
}

// +kubebuilder:object:root=true

// ProvisionedService is used to resolve a binding secret for a service resource
type ProvisionedService struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Status ProvisionedServiceStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ProvisionedServiceList contains a list of ProvisionedService
type ProvisionedServiceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []ProvisionedService `json:"items"`
}
