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

func UrlMapID(project string, key *meta.Key) *cloud.ResourceID {
	return &cloud.ResourceID{
		Resource:  "urlMaps",
		ProjectID: project,
		Key:       key,
	}
}

type MutableUrlMap = api.Resource[compute.UrlMap, alpha.UrlMap, beta.UrlMap]

func NewMutableUrlMap(project string, key *meta.Key) MutableUrlMap {
	id := UrlMapID(project, key)
	return api.NewResource[
		compute.UrlMap,
		alpha.UrlMap,
		beta.UrlMap,
	](id, &urlMapTypeTrait{})
}

type UrlMap = api.FrozenResource[compute.UrlMap, alpha.UrlMap, beta.UrlMap]

type UrlMapNode struct {
	nodeBase[compute.UrlMap, alpha.UrlMap, beta.UrlMap]
}

func newUrlMapNode(id *cloud.ResourceID) *UrlMapNode {
	ret := &UrlMapNode{}
	ret.init(id, &urlMapTypeTrait{})
	return ret
}

func UrlMapOutRefs(res UrlMap) ([]ResourceRef, error) {
	if res == nil {
		return nil, nil
	}

	var ret []ResourceRef
	obj, _ := res.ToGA()
	// DefaultService
	if obj.DefaultService != "" {
		id, err := cloud.ParseResourceURL(obj.DefaultService)
		if err != nil {
			return nil, fmt.Errorf("UrlMapNode DefaultService: %w", err)
		}
		ret = append(ret, ResourceRef{
			From: res.ResourceID(),
			Path: api.Path{}.Field("DefaultService"),
			To:   id,
		})
	}

	// TODO: rest of resource.

	return ret, nil
}

func (node *UrlMapNode) OutRefs() ([]ResourceRef, error) {
	return UrlMapOutRefs(node.resource)
}

func (node *UrlMapNode) Get(ctx context.Context, gcp cloud.Cloud) error {
	return genericGet[compute.UrlMap, alpha.UrlMap, beta.UrlMap](ctx, gcp, "UrlMap", &urlMapOps{}, &node.nodeBase)
}

func (node *UrlMapNode) NewEmptyNode() Node {
	ret := newUrlMapNode(node.ID())
	ret.initEmptyNode(node)
	return ret
}

func (node *UrlMapNode) Diff(gotNode Node) (*Action, error) {
	got, ok := gotNode.(*UrlMapNode)
	if !ok {
		return nil, fmt.Errorf("UrlMapNode: invalid type to Diff: %T", gotNode)
	}

	diff, err := got.resource.Diff(node.resource)
	if err != nil {
		return nil, fmt.Errorf("UrlMapNode: Diff %w", err)
	}

	if !diff.HasDiff() {
		// TODO: handle set labels with an update operation.
		return &Action{
			Operation: OpRecreate,
			Why:       "UrlMap needs to be recreated (no update method exists)",
			Diff:      diff,
		}, nil
	}

	return &Action{
		Operation: OpNothing,
		Why:       "No diff between got and want",
	}, nil
}

func (node *UrlMapNode) Actions() []exec.Action { return nil }

// https://cloud.google.com/compute/docs/reference/rest/v1/urlMaps
type urlMapTypeTrait struct {
	api.BaseTypeTrait[compute.UrlMap, alpha.UrlMap, beta.UrlMap]
}

func (*urlMapTypeTrait) FieldTraits(meta.Version) *api.FieldTraits {
	dt := api.NewFieldTraits()

	dt.OutputOnly(api.Path{}.Pointer().Field("CreationTimestamp"))
	dt.OutputOnly(api.Path{}.Pointer().Field("Id"))
	dt.OutputOnly(api.Path{}.Pointer().Field("Kind"))
	dt.OutputOnly(api.Path{}.Pointer().Field("Region"))
	dt.OutputOnly(api.Path{}.Pointer().Field("SelfLink"))
	dt.System(api.Path{}.Pointer().Field("Fingerprint"))

	return dt
}
