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

package projector

import (
	"context"

	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"

	servicebindingv1beta1 "github.com/servicebinding/runtime/apis/v1beta1"
)

type ServiceBindingProjector interface {
	// Project the service into the workload as defined by the ServiceBinding.
	Project(ctx context.Context, binding *servicebindingv1beta1.ServiceBinding, workload runtime.Object) error
	// Unproject the service from the workload as defined by the ServiceBinding.
	Unproject(ctx context.Context, binding *servicebindingv1beta1.ServiceBinding, workload runtime.Object) error
	// IsProjected returns true when the workload has been projected into by the binding
	IsProjected(ctx context.Context, binding *servicebindingv1beta1.ServiceBinding, workload runtime.Object) bool
}

type MappingSource interface {
	// LookupRESTMapping returns the RESTMapping for the workload type. The rest mapping contains a GroupVersionResource which can
	// be used to fetch the workload mapping.
	LookupRESTMapping(ctx context.Context, obj runtime.Object) (*meta.RESTMapping, error)

	// LookupWorkloadMapping the mapping template for the workload. Typically a ClusterWorkloadResourceMapping is defined for the
	//  workload's fully qualified resource `{resource}.{group}`.
	LookupWorkloadMapping(ctx context.Context, gvr schema.GroupVersionResource) (*servicebindingv1beta1.ClusterWorkloadResourceMappingSpec, error)
}
