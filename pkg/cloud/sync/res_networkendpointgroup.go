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

package sync

import (
	"context"
	"fmt"

	"github.com/GoogleCloudPlatform/k8s-cloud-provider/pkg/cloud"
	"github.com/GoogleCloudPlatform/k8s-cloud-provider/pkg/cloud/api"
	"github.com/GoogleCloudPlatform/k8s-cloud-provider/pkg/cloud/meta"
	"github.com/GoogleCloudPlatform/k8s-cloud-provider/pkg/cloud/sync/exec"

	alpha "google.golang.org/api/compute/v0.alpha"
	beta "google.golang.org/api/compute/v0.beta"
	"google.golang.org/api/compute/v1"
)

func NetworkEndpointGroupID(project string, key *meta.Key) *cloud.ResourceID {
	return &cloud.ResourceID{
		Resource:  "networkEndpointGroups",
		ProjectID: project,
		Key:       key,
	}
}

type MutableNetworkEndpointGroup = api.Resource[compute.NetworkEndpointGroup, alpha.NetworkEndpointGroup, beta.NetworkEndpointGroup]

func NewMutableNetworkEndpointGroup(project string, key *meta.Key) MutableNetworkEndpointGroup {
	id := NetworkEndpointGroupID(project, key)
	return api.NewResource[
		compute.NetworkEndpointGroup,
		alpha.NetworkEndpointGroup,
		beta.NetworkEndpointGroup,
	](id, &networkEndpointGroupTypeTrait{})
}

type NetworkEndpointGroup = api.FrozenResource[compute.NetworkEndpointGroup, alpha.NetworkEndpointGroup, beta.NetworkEndpointGroup]

type NetworkEndpointGroupNode struct {
	nodeBase[compute.NetworkEndpointGroup, alpha.NetworkEndpointGroup, beta.NetworkEndpointGroup]
}

func newNetworkEndpointGroupNode(id *cloud.ResourceID) *NetworkEndpointGroupNode {
	ret := &NetworkEndpointGroupNode{}
	ret.init(id, &networkEndpointGroupTypeTrait{})
	return ret
}

func NetworkEndpointGroupOutRefs(NetworkEndpointGroup) ([]ResourceRef, error) { return nil, nil }

func (node *NetworkEndpointGroupNode) OutRefs() ([]ResourceRef, error) {
	return NetworkEndpointGroupOutRefs(node.resource)
}

func (node *NetworkEndpointGroupNode) Get(ctx context.Context, gcp cloud.Cloud) error {
	return genericGet[compute.NetworkEndpointGroup, alpha.NetworkEndpointGroup, beta.NetworkEndpointGroup](ctx, gcp, "NetworkEndpointGroup", &networkEndpointGroupOps{}, &node.nodeBase)
}

func (node *NetworkEndpointGroupNode) NewEmptyNode() Node {
	ret := newNetworkEndpointGroupNode(node.ID())
	ret.initEmptyNode(node)
	return ret
}

func (node *NetworkEndpointGroupNode) Diff(gotNode Node) (*PlanDetails, error) {
	got, ok := gotNode.(*NetworkEndpointGroupNode)
	if !ok {
		return nil, fmt.Errorf("NetworkEndpointGroupNode: invalid type to Diff: %T", gotNode)
	}

	diff, err := got.resource.Diff(node.resource)
	if err != nil {
		return nil, fmt.Errorf("NetworkEndpointGroupNode: Diff %w", err)
	}

	if diff.HasDiff() {
		// TODO: handle set labels with an update operation.
		return &PlanDetails{
			Operation: OpRecreate,
			Why:       "NetworkEndpointGroup needs to be recreated (no update method exists)",
			Diff:      diff,
		}, nil
	}

	return &PlanDetails{
		Operation: OpNothing,
		Why:       "No diff between got and want",
	}, nil
}

func (node *NetworkEndpointGroupNode) Actions(got Node) ([]exec.Action, error) {
	op := node.LocalPlan().Op()

	switch op {
	case OpCreate:
		return opCreateActions[compute.NetworkEndpointGroup, alpha.NetworkEndpointGroup, beta.NetworkEndpointGroup](
			&networkEndpointGroupOps{}, node, node.resource)

	case OpDelete:
		return opDeleteActions[compute.NetworkEndpointGroup, alpha.NetworkEndpointGroup, beta.NetworkEndpointGroup](
			&networkEndpointGroupOps{}, got, node)

	case OpNothing:
		return []exec.Action{exec.NewExistsEventOnlyAction(node.ID())}, nil

	case OpRecreate:
		return opRecreateActions[compute.NetworkEndpointGroup, alpha.NetworkEndpointGroup, beta.NetworkEndpointGroup](
			&networkEndpointGroupOps{}, got, node, node.resource)

	case OpUpdate:
		// TODO
	}

	return nil, fmt.Errorf("NetworkEndpointGroupNode: invalid plan op %s", op)
}

// https://cloud.google.com/compute/docs/reference/rest/v1/networkEndpointGroups
type networkEndpointGroupTypeTrait struct {
	api.BaseTypeTrait[compute.NetworkEndpointGroup, alpha.NetworkEndpointGroup, beta.NetworkEndpointGroup]
}

func (*networkEndpointGroupTypeTrait) FieldTraits(meta.Version) *api.FieldTraits {
	dt := api.NewFieldTraits()
	// Built-ins
	dt.OutputOnly(api.Path{}.Pointer().Field("Fingerprint"))
	// [Output Only]
	dt.OutputOnly(api.Path{}.Pointer().Field("CreationTimestamp"))
	dt.OutputOnly(api.Path{}.Pointer().Field("Id"))
	dt.OutputOnly(api.Path{}.Pointer().Field("Kind"))
	dt.OutputOnly(api.Path{}.Pointer().Field("PscData"))
	dt.OutputOnly(api.Path{}.Pointer().Field("Region"))
	dt.OutputOnly(api.Path{}.Pointer().Field("SelfLink"))
	dt.OutputOnly(api.Path{}.Pointer().Field("Size"))
	dt.OutputOnly(api.Path{}.Pointer().Field("Zone"))

	// TODO: handle alpha/beta
	return dt
}
