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

func AddressID(project string, key *meta.Key) *cloud.ResourceID {
	return &cloud.ResourceID{
		Resource:  "addresses",
		ProjectID: project,
		Key:       key,
	}
}

type MutableAddress = api.Resource[compute.Address, alpha.Address, beta.Address]

func NewMutableAddress(project string, key *meta.Key) MutableAddress {
	id := AddressID(project, key)
	return api.NewResource[compute.Address, alpha.Address, beta.Address](id, &addressTypeTrait{})
}

type Address = api.FrozenResource[compute.Address, alpha.Address, beta.Address]

type AddressNode struct {
	nodeBase[compute.Address, alpha.Address, beta.Address]
}

func newAddressNode(id *cloud.ResourceID) *AddressNode {
	ret := &AddressNode{}
	ret.init(id, &addressTypeTrait{})
	return ret
}

func AddressOutRefs(Address) ([]ResourceRef, error) { return nil, nil }

func (node *AddressNode) OutRefs() ([]ResourceRef, error) { return AddressOutRefs(node.resource) }

func (node *AddressNode) Get(ctx context.Context, gcp cloud.Cloud) error {
	return genericGet[compute.Address, alpha.Address, beta.Address](ctx, gcp, "Address", &addressOps{}, &node.nodeBase)
}

func (node *AddressNode) NewEmptyNode() Node {
	ret := newAddressNode(node.ID())
	ret.initEmptyNode(node)
	return ret
}

func (node *AddressNode) Diff(gotNode Node) (*PlanDetails, error) {
	got, ok := gotNode.(*AddressNode)
	if !ok {
		return nil, fmt.Errorf("AddressNode: invalid type to Diff: %T", gotNode)
	}

	diff, err := got.resource.Diff(node.resource)
	if err != nil {
		return nil, fmt.Errorf("AddressNode: Diff %w", err)
	}

	if diff.HasDiff() {
		// TODO: handle set labels with an update operation.
		return &PlanDetails{
			Operation: OpRecreate,
			Why:       "Address needs to be recreated (no update method exists)",
			Diff:      diff,
		}, nil
	}

	return &PlanDetails{
		Operation: OpNothing,
		Why:       "No diff between got and want",
	}, nil
}

func (node *AddressNode) Actions(got Node) ([]exec.Action, error) {
	// TODO: got is needed for fingerprint

	op := node.LocalPlan().Op()

	switch op {
	case OpCreate:
		return opCreateActions[compute.Address, alpha.Address, beta.Address](&addressOps{}, node, node.resource)

	case OpDelete:
		return opDeleteActions[compute.Address, alpha.Address, beta.Address](&addressOps{}, got, node)

	case OpNothing:
		return []exec.Action{exec.NewExistsEventOnlyAction(node.ID())}, nil

	case OpRecreate:
		return opRecreateActions[compute.Address, alpha.Address, beta.Address](&addressOps{}, got, node, node.resource)

	case OpUpdate:
		// TODO
	}

	return nil, fmt.Errorf("AddressNode: invalid plan op %s", op)
}

// See https://cloud.google.com/compute/docs/reference/rest/v1/addresses
type addressTypeTrait struct {
	api.BaseTypeTrait[compute.Address, alpha.Address, beta.Address]
}

func (*addressTypeTrait) FieldTraits(meta.Version) *api.FieldTraits {
	dt := api.NewFieldTraits()
	// Built-ins
	dt.OutputOnly(api.Path{}.Pointer().Field("Fingerprint"))
	// [Output Only]
	dt.OutputOnly(api.Path{}.Pointer().Field("Kind"))
	dt.OutputOnly(api.Path{}.Pointer().Field("Id"))
	dt.OutputOnly(api.Path{}.Pointer().Field("CreationTimestamp"))
	dt.OutputOnly(api.Path{}.Pointer().Field("Status"))
	dt.OutputOnly(api.Path{}.Pointer().Field("Region"))
	dt.OutputOnly(api.Path{}.Pointer().Field("SelfLink"))
	dt.OutputOnly(api.Path{}.Pointer().Field("Users"))

	// TODO: handle alpha/beta

	return dt
}
