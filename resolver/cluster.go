/*
Copyright 2021 the original author or authors.

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

package resolver

import (
	"context"
	"fmt"

	rduck "github.com/vmware-labs/reconciler-runtime/duck"
	corev1 "k8s.io/api/core/v1"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"

	"github.com/servicebinding/runtime/apis/duck"
	servicebindingv1beta1 "github.com/servicebinding/runtime/apis/v1beta1"
)

// New creates a new resolver backed by a controller-runtime client
func New(client client.Client) Resolver {
	return &clusterResolver{
		client: rduck.NewDuckAwareClientWrapper(client),
	}
}

type clusterResolver struct {
	client client.Client
}

func (m *clusterResolver) LookupRESTMapping(ctx context.Context, obj runtime.Object) (*meta.RESTMapping, error) {
	gvk, err := apiutil.GVKForObject(obj, m.client.Scheme())
	if err != nil {
		return nil, err
	}
	rm, err := m.client.RESTMapper().RESTMapping(gvk.GroupKind(), gvk.Version)
	if err != nil {
		return nil, err
	}
	return rm, nil
}

func (m *clusterResolver) LookupWorkloadMapping(ctx context.Context, gvr schema.GroupVersionResource) (*servicebindingv1beta1.ClusterWorkloadResourceMappingSpec, error) {
	wrm := &servicebindingv1beta1.ClusterWorkloadResourceMapping{}

	if err := m.client.Get(ctx, types.NamespacedName{Name: fmt.Sprintf("%s.%s", gvr.Resource, gvr.Group)}, wrm); err != nil {
		if !apierrs.IsNotFound(err) {
			return nil, err
		}
		wrm.Spec = servicebindingv1beta1.ClusterWorkloadResourceMappingSpec{
			Versions: []servicebindingv1beta1.ClusterWorkloadResourceMappingTemplate{
				{
					Version: "*",
				},
			},
		}
	}

	for i := range wrm.Spec.Versions {
		wrm.Spec.Versions[i].Default()
	}

	return &wrm.Spec, nil
}

func (r *clusterResolver) LookupBindingSecret(ctx context.Context, serviceRef corev1.ObjectReference) (string, error) {
	if serviceRef.APIVersion == "v1" && serviceRef.Kind == "Secret" {
		// direct secret reference
		return serviceRef.Name, nil
	}

	service := &duck.ProvisionedService{
		TypeMeta: metav1.TypeMeta{
			APIVersion: serviceRef.APIVersion,
			Kind:       serviceRef.Kind,
		},
	}
	key := client.ObjectKey{
		Namespace: serviceRef.Namespace,
		Name:      serviceRef.Name,
	}
	if err := r.client.Get(ctx, key, service); err != nil {
		return "", err
	}
	return service.Status.Binding.Name, nil
}

func (r *clusterResolver) LookupWorkloads(ctx context.Context, workloadRef corev1.ObjectReference, selector *metav1.LabelSelector) ([]runtime.Object, error) {
	if workloadRef.Name != "" {
		workload, err := r.lookupWorkload(ctx, workloadRef)
		if err != nil {
			return nil, err
		}
		return []runtime.Object{workload}, nil
	}
	return r.lookupWorkloads(ctx, workloadRef, selector)
}

func (r *clusterResolver) lookupWorkload(ctx context.Context, workloadRef corev1.ObjectReference) (runtime.Object, error) {
	workload := &unstructured.Unstructured{}
	workload.SetAPIVersion(workloadRef.APIVersion)
	workload.SetKind(workloadRef.Kind)
	if err := r.client.Get(ctx, client.ObjectKey{Namespace: workloadRef.Namespace, Name: workloadRef.Name}, workload); err != nil {
		return nil, err
	}
	return workload, nil
}

func (r *clusterResolver) lookupWorkloads(ctx context.Context, workloadRef corev1.ObjectReference, selector *metav1.LabelSelector) ([]runtime.Object, error) {
	workloads := &unstructured.UnstructuredList{}
	workloads.SetAPIVersion(workloadRef.APIVersion)
	workloads.SetKind(fmt.Sprintf("%sList", workloadRef.Kind))
	ls, err := metav1.LabelSelectorAsSelector(selector)
	if err != nil {
		return nil, err
	}
	if err := r.client.List(ctx, workloads, client.InNamespace(workloadRef.Namespace), client.MatchingLabelsSelector{Selector: ls}); err != nil {
		return nil, err
	}

	// coerce to []runtime.Object
	result := make([]runtime.Object, len(workloads.Items))
	for i := range workloads.Items {
		result[i] = &workloads.Items[i]
	}
	return result, nil
}
