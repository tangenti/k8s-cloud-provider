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

func TargetHttpProxyID(project string, key *meta.Key) *cloud.ResourceID {
	return &cloud.ResourceID{
		Resource:  "targetHttpProxies",
		ProjectID: project,
		Key:       key,
	}
}

type MutableTargetHttpProxy = api.Resource[compute.TargetHttpProxy, alpha.TargetHttpProxy, beta.TargetHttpProxy]

func NewMutableTargetHttpProxy(project string, key *meta.Key) MutableTargetHttpProxy {
	id := TargetHttpProxyID(project, key)
	return api.NewResource[
		compute.TargetHttpProxy,
		alpha.TargetHttpProxy,
		beta.TargetHttpProxy,
	](id, &targetHttpProxyTypeTrait{})
}

type TargetHttpProxy = api.FrozenResource[compute.TargetHttpProxy, alpha.TargetHttpProxy, beta.TargetHttpProxy]

type TargetHttpProxyNode struct {
	nodeBase[compute.TargetHttpProxy, alpha.TargetHttpProxy, beta.TargetHttpProxy]
}

func newTargetHttpProxyNode(id *cloud.ResourceID) *TargetHttpProxyNode {
	ret := &TargetHttpProxyNode{}
	ret.init(id, &targetHttpProxyTypeTrait{})
	return ret
}

func TargetHttpProxyOutRefs(res TargetHttpProxy) ([]ResourceRef, error) {
	if res == nil {
		return nil, nil
	}

	var ret []ResourceRef
	obj, _ := res.ToGA()

	if obj.UrlMap != "" {
		id, err := cloud.ParseResourceURL(obj.UrlMap)
		if err != nil {
			return nil, fmt.Errorf("TargetHttpProxyNode: %w", err)
		}
		ret = append(ret, ResourceRef{
			From: res.ResourceID(),
			Path: api.Path{}.Field("UrlMap"),
			To:   id,
		})
	}

	return ret, nil
}

func (node *TargetHttpProxyNode) OutRefs() ([]ResourceRef, error) {
	return TargetHttpProxyOutRefs(node.resource)
}

func (node *TargetHttpProxyNode) Get(ctx context.Context, gcp cloud.Cloud) error {
	return genericGet[compute.TargetHttpProxy, alpha.TargetHttpProxy, beta.TargetHttpProxy](ctx, gcp, "TargetHttpProxy", &targetHttpProxyOps{}, &node.nodeBase)
}

func (node *TargetHttpProxyNode) NewEmptyNode() Node {
	ret := newTargetHttpProxyNode(node.ID())
	ret.initEmptyNode(node)
	return ret
}

func (node *TargetHttpProxyNode) Diff(gotNode Node) (*PlanDetails, error) {
	got, ok := gotNode.(*TargetHttpProxyNode)
	if !ok {
		return nil, fmt.Errorf("TargetHttpProxyNode: invalid type to Diff: %T", gotNode)
	}

	diff, err := got.resource.Diff(node.resource)
	if err != nil {
		return nil, fmt.Errorf("TargetHttpProxyNode: Diff %w", err)
	}

	if diff.HasDiff() {
		// TODO: handle set labels with an update operation.
		return &PlanDetails{
			Operation: OpRecreate,
			Why:       "TargetHttpProxy needs to be recreated (no update method exists)",
			Diff:      diff,
		}, nil
	}

	return &PlanDetails{
		Operation: OpNothing,
		Why:       "No diff between got and want",
	}, nil
}

func (node *TargetHttpProxyNode) Actions(got Node) ([]exec.Action, error) {
	op := node.LocalPlan().Op()

	switch op {
	case OpCreate:
		return opCreateActions[compute.TargetHttpProxy, alpha.TargetHttpProxy, beta.TargetHttpProxy](&targetHttpProxyOps{}, node, node.resource)

	case OpDelete:
		return opDeleteActions[compute.TargetHttpProxy, alpha.TargetHttpProxy, beta.TargetHttpProxy](&targetHttpProxyOps{}, got, node)

	case OpNothing:
		return []exec.Action{exec.NewExistsEventOnlyAction(node.ID())}, nil

	case OpRecreate:
		return opRecreateActions[compute.TargetHttpProxy, alpha.TargetHttpProxy, beta.TargetHttpProxy](&targetHttpProxyOps{}, got, node, node.resource)

	case OpUpdate:
		// TODO
	}

	return nil, fmt.Errorf("TargetHttpProxyNode: invalid plan op %s", op)
}

// https://cloud.google.com/compute/docs/reference/rest/v1/targetHttpProxies
type targetHttpProxyTypeTrait struct {
	api.BaseTypeTrait[compute.TargetHttpProxy, alpha.TargetHttpProxy, beta.TargetHttpProxy]
}

func (*targetHttpProxyTypeTrait) FieldTraits(meta.Version) *api.FieldTraits {
	dt := api.NewFieldTraits()
	// Built-ins
	dt.OutputOnly(api.Path{}.Pointer().Field("Fingerprint"))
	// [Output Only]
	dt.OutputOnly(api.Path{}.Pointer().Field("CreationTimestamp"))
	dt.OutputOnly(api.Path{}.Pointer().Field("Id"))
	dt.OutputOnly(api.Path{}.Pointer().Field("Kind"))
	dt.OutputOnly(api.Path{}.Pointer().Field("SelfLink"))
	// TODO: finish me
	// TODO: handle alpha/beta
	return dt
}
