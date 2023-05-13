/*
Copyright 2018 Google LLC

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

https://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package cloud

import (
	"context"

	"github.com/GoogleCloudPlatform/k8s-cloud-provider/pkg/cloud/meta"
	"k8s.io/klog/v2"
)

// ProjectRouter routes service calls to the appropriate GCE project.
type ProjectRouter interface {
	// ProjectID returns the project ID (non-numeric) to be used for a call
	// to an API (version,service). Example tuples: ("ga", "ForwardingRules"),
	// ("alpha", "GlobalAddresses").
	//
	// This allows for plumbing different service calls to the appropriate
	// project, for instance, networking services to a separate project
	// than instance management.
	ProjectID(ctx context.Context, version meta.Version, service string) string
}

// SingleProjectRouter routes all service calls to the same project ID.
type SingleProjectRouter struct {
	ID string
}

// ProjectID returns the project ID to be used for a call to the API.
func (r *SingleProjectRouter) ProjectID(ctx context.Context, version meta.Version, service string) string {
	return r.ID
}

// ContextProjectRouter uses a context key to provide an optional
// project routing.
type ContextProjectRouter struct {
	DefaultProject string
}

// ProjectID implements ProjectRouter.
func (r *ContextProjectRouter) ProjectID(ctx context.Context, version meta.Version, service string) string {
	v := ctx.Value(projectContextKey("projectContext"))
	if v == nil {
		return r.DefaultProject
	}
	project, ok := v.(string)
	if !ok {
		klog.Errorf("invalid value type for projectContextKey: %T, value was %v", v, v)
		return r.DefaultProject
	}
	return project
}

type projectContextKey string

// WithProjectKey returns a new Context with the project context added.
func WithProjectKey(ctx context.Context, project string) context.Context {
	return context.WithValue(ctx, projectContextKey("projectContext"), project)
}
