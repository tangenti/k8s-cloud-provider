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

func BackendServiceID(project string, key *meta.Key) *cloud.ResourceID {
	return &cloud.ResourceID{
		Resource:  "backendServices",
		ProjectID: project,
		Key:       key,
	}
}

type MutableBackendService = api.Resource[compute.BackendService, alpha.BackendService, beta.BackendService]

func NewMutableBackendService(project string, key *meta.Key) MutableBackendService {
	id := BackendServiceID(project, key)
	return api.NewResource[
		compute.BackendService,
		alpha.BackendService,
		beta.BackendService,
	](id, &backendServiceTypeTrait{})
}

type BackendService = api.FrozenResource[compute.BackendService, alpha.BackendService, beta.BackendService]

type BackendServiceNode struct {
	nodeBase[compute.BackendService, alpha.BackendService, beta.BackendService]
}

func newBackendServiceNode(id *cloud.ResourceID) *BackendServiceNode {
	ret := &BackendServiceNode{}
	ret.init(id, &backendServiceTypeTrait{})
	return ret
}

func BackendServiceOutRefs(res BackendService) ([]ResourceRef, error) {
	if res == nil {
		return nil, nil
	}

	obj, _ := res.ToGA()

	var ret []ResourceRef

	// Backends[].Group
	for idx, backend := range obj.Backends {
		id, err := cloud.ParseResourceURL(backend.Group)
		if err != nil {
			return nil, fmt.Errorf("BackendServiceNode Group: %w", err)
		}
		ret = append(ret, ResourceRef{
			From: res.ResourceID(),
			Path: api.Path{}.Field("Backends").Index(idx).Field("Group"),
			To:   id,
		})
	}

	// Healthchecks[]
	for idx, hc := range obj.HealthChecks {
		id, err := cloud.ParseResourceURL(hc)
		if err != nil {
			return nil, fmt.Errorf("BackendServiceNode HealthChecks: %w", err)
		}
		ret = append(ret, ResourceRef{
			From: res.ResourceID(),
			Path: api.Path{}.Field("HealthChecks").Index(idx),
			To:   id,
		})
	}

	// SecurityPolicy
	if obj.SecurityPolicy != "" {
		id, err := cloud.ParseResourceURL(obj.SecurityPolicy)
		if err != nil {
			return nil, fmt.Errorf("BackendServiceNode SecurityPolicy: %w", err)
		}
		ret = append(ret, ResourceRef{
			From: res.ResourceID(),
			Path: api.Path{}.Field("SecurityPolicy"),
			To:   id,
		})
	}

	// EdgeSecurityPolicy
	if obj.EdgeSecurityPolicy != "" {
		id, err := cloud.ParseResourceURL(obj.EdgeSecurityPolicy)
		if err != nil {
			return nil, fmt.Errorf("BackendServiceNode SecurityPolicy: %w", err)
		}
		ret = append(ret, ResourceRef{
			From: res.ResourceID(),
			Path: api.Path{}.Field("SecurityPolicy"),
			To:   id,
		})
	}

	return ret, nil
}

func (node *BackendServiceNode) OutRefs() ([]ResourceRef, error) {
	return BackendServiceOutRefs(node.resource)
}

func (node *BackendServiceNode) Get(ctx context.Context, gcp cloud.Cloud) error {
	return genericGet[compute.BackendService, alpha.BackendService, beta.BackendService](
		ctx, gcp, "BackendService", &backendServiceOps{}, &node.nodeBase)
}

func (node *BackendServiceNode) NewEmptyNode() Node {
	ret := newBackendServiceNode(node.ID())
	ret.initEmptyNode(node)
	return ret
}

func (node *BackendServiceNode) Diff(gotNode Node) (*Action, error) {
	got, ok := gotNode.(*BackendServiceNode)
	if !ok {
		return nil, fmt.Errorf("BackendServiceNode: invalid type to Diff: %T", gotNode)
	}

	diff, err := got.resource.Diff(node.resource)
	if err != nil {
		return nil, fmt.Errorf("BackendServiceNode: Diff %w", err)
	}

	if diff.HasDiff() {
		// TODO: XXX
		return &Action{
			Operation: OpRecreate,
			Why:       "BackendService needs to be recreated (no update method exists)",
			Diff:      diff,
		}, nil
	}

	return &Action{
		Operation: OpNothing,
		Why:       "No diff between got and want",
	}, nil
}

func (node *BackendServiceNode) Actions(got Node) ([]exec.Action, error) {
	op := node.LocalPlan().Op()

	switch op {
	case OpCreate:
		return opCreateActions[compute.BackendService, alpha.BackendService, beta.BackendService](&backendServiceOps{}, node, node.resource)

	case OpDelete:
		return opDeleteActions[compute.BackendService, alpha.BackendService, beta.BackendService](&backendServiceOps{}, got, node)

	case OpNothing:
		return []exec.Action{exec.NewExistsEventOnlyAction(node.ID())}, nil

	case OpRecreate:
		return opRecreateActions[compute.BackendService, alpha.BackendService, beta.BackendService](&backendServiceOps{}, got, node, node.resource)

	case OpUpdate:
		// TODO
	}

	return nil, fmt.Errorf("BackendServiceNode: invalid plan op %s", op)
}

// https://cloud.google.com/compute/docs/reference/rest/v1/backendServices
type backendServiceTypeTrait struct {
	api.BaseTypeTrait[compute.BackendService, alpha.BackendService, beta.BackendService]
}

func (*backendServiceTypeTrait) FieldTraits(meta.Version) *api.FieldTraits {
	dt := api.NewFieldTraits()
	// Built-ins
	dt.OutputOnly(api.Path{}.Pointer().Field("Fingerprint"))

	// [Output Only]
	dt.OutputOnly(api.Path{}.Pointer().Field("CreationTimestamp"))
	dt.OutputOnly(api.Path{}.Pointer().Field("EdgeSecurityPolicy"))
	dt.OutputOnly(api.Path{}.Pointer().Field("Id"))
	dt.OutputOnly(api.Path{}.Pointer().Field("Kind"))
	dt.OutputOnly(api.Path{}.Pointer().Field("Region"))
	dt.OutputOnly(api.Path{}.Pointer().Field("SecurityPolicy"))
	dt.OutputOnly(api.Path{}.Pointer().Field("SelfLink"))

	dt.OutputOnly(api.Path{}.Pointer().Field("Iap").Field("Oauth2ClientSecretSha256"))
	dt.OutputOnly(api.Path{}.Pointer().Field("CdnPolicy").Field("SignedUrlKeyNames"))

	// TODO: finish me
	// TODO: handle alpha/beta

	return dt
}
