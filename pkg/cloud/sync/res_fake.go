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
)

// fakeResource is a resource used only for testing.
type fakeResource struct {
	Name            string
	Dependencies    []string
	NullFields      []string
	ForceSendFields []string
}

func fakeID(project string, key *meta.Key) *cloud.ResourceID {
	return &cloud.ResourceID{
		Resource:  "fakes",
		ProjectID: project,
		Key:       key,
	}
}

type mutableFake = api.Resource[fakeResource, fakeResource, fakeResource]

func newMutableFake(project string, key *meta.Key) mutableFake {
	res := &cloud.ResourceID{
		Resource:  "fakes",
		ProjectID: project,
		Key:       key,
	}
	return api.NewResource[fakeResource, fakeResource, fakeResource](res, &fakeTypeTrait{})
}

type fake = api.FrozenResource[fakeResource, fakeResource, fakeResource]

type fakeNode struct {
	nodeBase[fakeResource, fakeResource, fakeResource]
}

func newFakeNode(id *cloud.ResourceID) *fakeNode {
	return &fakeNode{
		nodeBase: nodeBase[fakeResource, fakeResource, fakeResource]{
			id:        id,
			typeTrait: &fakeTypeTrait{},
			state:     NodeUnknown,
			ownership: OwnershipUnknown,
		},
	}
}

func (node *fakeNode) OutRefs() ([]ResourceRef, error) {
	var ret []ResourceRef
	r, _ := node.resource.ToGA()
	for _, dep := range r.Dependencies {
		id, err := cloud.ParseResourceURL(dep)
		if err != nil {
			return nil, err
		}
		ret = append(ret, ResourceRef{
			From: node.ID(),
			Path: api.Path{}.Field("Dependencies"),
			To:   id,
		})
	}
	return ret, nil
}

func (node *fakeNode) NewEmptyNode() Node {
	ret := newFakeNode(node.ID())
	ret.initEmptyNode(node)
	return ret
}

func (node *fakeNode) Diff(gotNode Node) (*PlanAction, error) {
	got, ok := gotNode.(*fakeNode)
	if !ok {
		return nil, fmt.Errorf("FakeNode: invalid type to Diff: %T", gotNode)
	}
	diff, err := got.resource.Diff(node.resource)
	if err != nil {
		return nil, fmt.Errorf("FakeNode: invalid type to Diff: %T", gotNode)
	}
	if !diff.HasDiff() {
		return &PlanAction{
			Operation: OpNothing,
			Why:       "No diff between got and want",
		}, nil
	}

	// TODO: handle set labels with an update operation.

	return &PlanAction{
		Operation: OpRecreate,
		Why:       "Fake needs to be recreated (no update method exists)",
	}, nil
}

func (*fakeNode) Get(context.Context, cloud.Cloud) error  { return nil }
func (*fakeNode) Sync(context.Context, cloud.Cloud) error { return nil }
func (*fakeNode) GenerateLocalPlan() error                { return nil }
func (*fakeNode) Actions(got Node) ([]exec.Action, error) { return nil, nil }

type fakeTypeTrait struct {
	api.BaseTypeTrait[fakeResource, fakeResource, fakeResource]
}
