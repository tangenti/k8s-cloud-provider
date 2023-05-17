/*
Copyright 2023 Google LLC

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

package resgraph

import (
	"context"
	"fmt"

	"github.com/GoogleCloudPlatform/k8s-cloud-provider/pkg/cloud"
	"github.com/GoogleCloudPlatform/k8s-cloud-provider/pkg/cloud/api"
	"github.com/GoogleCloudPlatform/k8s-cloud-provider/pkg/cloud/meta"
	"github.com/GoogleCloudPlatform/k8s-cloud-provider/pkg/cloud/resgraph/exec"

	alpha "google.golang.org/api/compute/v0.alpha"
	beta "google.golang.org/api/compute/v0.beta"
	"google.golang.org/api/compute/v1"
)

func HealthCheckID(project string, key *meta.Key) *cloud.ResourceID {
	return &cloud.ResourceID{
		Resource:  "healthChecks",
		ProjectID: project,
		Key:       key,
	}
}

type MutableHealthCheck = api.Resource[compute.HealthCheck, alpha.HealthCheck, beta.HealthCheck]

func NewMutableHealthCheck(project string, key *meta.Key) MutableHealthCheck {
	id := HealthCheckID(project, key)
	return api.NewResource[
		compute.HealthCheck,
		alpha.HealthCheck,
		beta.HealthCheck,
	](id, &healthCheckTypeTrait{})
}

type HealthCheck = api.FrozenResource[compute.HealthCheck, alpha.HealthCheck, beta.HealthCheck]

type HealthCheckNode struct {
	nodeBase[compute.HealthCheck, alpha.HealthCheck, beta.HealthCheck]
}

func newHealthCheckNode(id *cloud.ResourceID) *HealthCheckNode {
	ret := &HealthCheckNode{}
	ret.init(id, &healthCheckTypeTrait{})
	return ret
}

func HealthCheckOutRefs(HealthCheck) ([]ResourceRef, error) { return nil, nil }

func (node *HealthCheckNode) OutRefs() ([]ResourceRef, error) {
	return HealthCheckOutRefs(node.resource)
}

func (node *HealthCheckNode) Get(ctx context.Context, gcp cloud.Cloud) error {
	return genericGet[compute.HealthCheck, alpha.HealthCheck, beta.HealthCheck](ctx, gcp, "HealthCheck", &healthCheckOps{}, &node.nodeBase)
}

func (node *HealthCheckNode) NewEmptyNode() Node {
	ret := newHealthCheckNode(node.ID())
	ret.initEmptyNode(node)
	return ret
}

func (node *HealthCheckNode) Diff(gotNode Node) (*PlanDetails, error) {
	got, ok := gotNode.(*HealthCheckNode)
	if !ok {
		return nil, fmt.Errorf("HealthCheckNode: invalid type to Diff: %T", gotNode)
	}

	diff, err := got.resource.Diff(node.resource)
	if err != nil {
		return nil, fmt.Errorf("HealthCheckNode: Diff %w", err)
	}

	if diff.HasDiff() {
		// TODO: handle set labels with an update operation.
		return &PlanDetails{
			Operation: OpRecreate,
			Why:       "HealthCheck needs to be recreated (no update method exists)",
			Diff:      diff,
		}, nil
	}

	return &PlanDetails{
		Operation: OpNothing,
		Why:       "No diff between got and want",
	}, nil
}

func (node *HealthCheckNode) Actions(got Node) ([]exec.Action, error) {
	op := node.LocalPlan().Op()

	switch op {
	case OpCreate:
		return opCreateActions[compute.HealthCheck, alpha.HealthCheck, beta.HealthCheck](&healthCheckOps{}, node, node.resource)

	case OpDelete:
		return opDeleteActions[compute.HealthCheck, alpha.HealthCheck, beta.HealthCheck](&healthCheckOps{}, got, node)

	case OpNothing:
		return []exec.Action{exec.NewExistsEventOnlyAction(node.ID())}, nil

	case OpRecreate:
		return opRecreateActions[compute.HealthCheck, alpha.HealthCheck, beta.HealthCheck](&healthCheckOps{}, got, node, node.resource)

	case OpUpdate:
		// TODO
	}

	return nil, fmt.Errorf("HealthCheckNode: invalid plan op %s", op)
}

// https://cloud.google.com/compute/docs/reference/rest/v1/HealthChecks
type healthCheckTypeTrait struct {
	api.BaseTypeTrait[compute.HealthCheck, alpha.HealthCheck, beta.HealthCheck]
}

func (*healthCheckTypeTrait) FieldTraits(meta.Version) *api.FieldTraits {
	dt := api.NewFieldTraits()
	// Built-ins
	dt.OutputOnly(api.Path{}.Pointer().Field("Fingerprint"))

	// [Output Only]
	dt.OutputOnly(api.Path{}.Pointer().Field("CreationTimestamp"))
	dt.OutputOnly(api.Path{}.Pointer().Field("Id"))
	dt.OutputOnly(api.Path{}.Pointer().Field("Kind"))
	dt.OutputOnly(api.Path{}.Pointer().Field("Region"))
	dt.OutputOnly(api.Path{}.Pointer().Field("SelfLink"))

	// TODO: handle alpha/beta

	return dt
}
