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
	"net"

	"github.com/GoogleCloudPlatform/k8s-cloud-provider/pkg/cloud"
	"github.com/GoogleCloudPlatform/k8s-cloud-provider/pkg/cloud/api"
	"github.com/GoogleCloudPlatform/k8s-cloud-provider/pkg/cloud/meta"
	"github.com/GoogleCloudPlatform/k8s-cloud-provider/pkg/cloud/sync/exec"
	"k8s.io/klog/v2"

	alpha "google.golang.org/api/compute/v0.alpha"
	beta "google.golang.org/api/compute/v0.beta"
	"google.golang.org/api/compute/v1"
)

func ForwardingRuleID(project string, key *meta.Key) *cloud.ResourceID {
	return &cloud.ResourceID{
		Resource:  "forwardingRules",
		ProjectID: project,
		Key:       key,
	}
}

type MutableForwardingRule = api.Resource[compute.ForwardingRule, alpha.ForwardingRule, beta.ForwardingRule]

func NewMutableForwardingRule(project string, key *meta.Key) MutableForwardingRule {
	id := ForwardingRuleID(project, key)
	return api.NewResource[
		compute.ForwardingRule,
		alpha.ForwardingRule,
		beta.ForwardingRule,
	](id, &forwardingRuleTypeTrait{})
}

type ForwardingRule = api.FrozenResource[compute.ForwardingRule, alpha.ForwardingRule, beta.ForwardingRule]

type ForwardingRuleNode struct {
	nodeBase[compute.ForwardingRule, alpha.ForwardingRule, beta.ForwardingRule]
}

func newForwardingRuleNode(id *cloud.ResourceID) *ForwardingRuleNode {
	ret := &ForwardingRuleNode{}
	ret.init(id, &forwardingRuleTypeTrait{})
	return ret
}

func ForwardingRuleOutRefs(res ForwardingRule) ([]ResourceRef, error) {
	if res == nil {
		return nil, nil
	}

	var ret []ResourceRef
	obj, err := res.ToGA()

	klog.Infof("res.ToGA = %+v, %v", obj, err)

	// IPAddress
	if obj.IPAddress != "" {
		// Handle numeric IP address.
		if ip := net.ParseIP(obj.IPAddress); ip != nil {
			// This is emphemeral, so is not a real resource.
		} else {
			id, err := cloud.ParseResourceURL(obj.IPAddress)
			klog.Infof("IPAddress = %+v, %v", id, err)
			if err != nil {
				return nil, fmt.Errorf("ForwardingRuleNode IPAddress: %w", err)
			}
			ret = append(ret, ResourceRef{
				From: res.ResourceID(),
				Path: api.Path{}.Field("IPAddress"),
				To:   id,
			})
		}
	}

	// Target
	if obj.Target != "" {
		id, err := cloud.ParseResourceURL(obj.Target)
		klog.Infof("Target = %+v, %v", id, err)
		if err != nil {
			return nil, fmt.Errorf("ForwardingRuleNode Target: %w", err)
		}
		ret = append(ret, ResourceRef{
			From: res.ResourceID(),
			Path: api.Path{}.Field("Target"),
			To:   id,
		})
	}

	return ret, nil
}

func (node *ForwardingRuleNode) OutRefs() ([]ResourceRef, error) {
	return ForwardingRuleOutRefs(node.resource)
}

func (node *ForwardingRuleNode) Get(ctx context.Context, gcp cloud.Cloud) error {
	return genericGet[compute.ForwardingRule, alpha.ForwardingRule, beta.ForwardingRule](ctx, gcp, "ForwardingRule", &forwardingRuleOps{}, &node.nodeBase)
}

func (node *ForwardingRuleNode) NewEmptyNode() Node {
	ret := newForwardingRuleNode(node.ID())
	ret.initEmptyNode(node)
	return ret
}

func (node *ForwardingRuleNode) Diff(gotNode Node) (*PlanDetails, error) {
	got, ok := gotNode.(*ForwardingRuleNode)
	if !ok {
		return nil, fmt.Errorf("ForwardingRuleNode: invalid type to Diff: %T", gotNode)
	}

	diff, err := got.resource.Diff(node.resource)
	if err != nil {
		return nil, fmt.Errorf("ForwardingRuleNode: Diff %w", err)
	}

	if !diff.HasDiff() {
		// TODO:
		return &PlanDetails{
			Operation: OpRecreate,
			Why:       "ForwardingRule needs to be recreated (no update method exists)",
			Diff:      diff,
		}, nil
	}

	return &PlanDetails{
		Operation: OpNothing,
		Why:       "No diff between got and want",
	}, nil
}

func (node *ForwardingRuleNode) Actions(got Node) ([]exec.Action, error) {
	op := node.LocalPlan().Op()

	switch op {
	case OpCreate:
		return opCreateActions[compute.ForwardingRule, alpha.ForwardingRule, beta.ForwardingRule](&forwardingRuleOps{}, node, node.resource)

	case OpDelete:
		return opDeleteActions[compute.ForwardingRule, alpha.ForwardingRule, beta.ForwardingRule](&forwardingRuleOps{}, got, node)

	case OpNothing:
		return []exec.Action{exec.NewExistsEventOnlyAction(node.ID())}, nil

	case OpRecreate:
		return opRecreateActions[compute.ForwardingRule, alpha.ForwardingRule, beta.ForwardingRule](&forwardingRuleOps{}, got, node, node.resource)

	case OpUpdate:
		// TODO
	}

	return nil, fmt.Errorf("ForwardingRule: invalid plan op %s", op)
}

// https://cloud.google.com/compute/docs/reference/rest/beta/forwardingRules
type forwardingRuleTypeTrait struct {
	api.BaseTypeTrait[compute.ForwardingRule, alpha.ForwardingRule, beta.ForwardingRule]
}

func (*forwardingRuleTypeTrait) FieldTraits(meta.Version) *api.FieldTraits {
	dt := api.NewFieldTraits()
	// Built-ins
	dt.OutputOnly(api.Path{}.Pointer().Field("Fingerprint"))
	// [Output Only]
	dt.OutputOnly(api.Path{}.Pointer().Field("Kind"))
	dt.OutputOnly(api.Path{}.Pointer().Field("Id"))
	dt.OutputOnly(api.Path{}.Pointer().Field("CreationTimestamp"))
	dt.OutputOnly(api.Path{}.Pointer().Field("Region"))
	dt.OutputOnly(api.Path{}.Pointer().Field("SelfLink"))
	dt.OutputOnly(api.Path{}.Pointer().Field("ServiceName"))
	dt.OutputOnly(api.Path{}.Pointer().Field("PscConnectionId"))
	dt.OutputOnly(api.Path{}.Pointer().Field("PscConnectionStatus")) // Not documented
	dt.OutputOnly(api.Path{}.Pointer().Field("BaseForwardingRule"))

	// TODO: handle alpha/beta

	return dt
}
