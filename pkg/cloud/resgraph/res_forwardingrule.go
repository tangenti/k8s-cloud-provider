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
	"net"

	"github.com/GoogleCloudPlatform/k8s-cloud-provider/pkg/cloud"
	"github.com/GoogleCloudPlatform/k8s-cloud-provider/pkg/cloud/api"
	"github.com/GoogleCloudPlatform/k8s-cloud-provider/pkg/cloud/meta"
	"github.com/GoogleCloudPlatform/k8s-cloud-provider/pkg/cloud/resgraph/exec"
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

	if diff.HasDiff() {
		var changed struct{ target, labels, other bool }
		for _, item := range diff.Items {
			switch {
			case api.Path{}.Pointer().Field("Target").Equal(item.Path):
				changed.target = true
			case api.Path{}.Pointer().Field("Labels").Equal(item.Path):
				changed.labels = true
			default:
				klog.Errorf("XXX other = %v", item)
				changed.other = true
			}
		}

		if !changed.other {
			return &PlanDetails{
				Operation: OpUpdate,
				Why:       fmt.Sprintf("update in place (changed=%+v)", changed),
				Diff:      diff,
			}, nil
		}

		return &PlanDetails{
			Operation: OpRecreate,
			Why:       "needs to be recreated",
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
		return node.createActions(got)

	case OpDelete:
		return opDeleteActions[compute.ForwardingRule, alpha.ForwardingRule, beta.ForwardingRule](&forwardingRuleOps{}, got, node)

	case OpNothing:
		return []exec.Action{exec.NewExistsEventOnlyAction(node.ID())}, nil

	case OpRecreate:
		return opRecreateActions[compute.ForwardingRule, alpha.ForwardingRule, beta.ForwardingRule](&forwardingRuleOps{}, got, node, node.resource)

	case OpUpdate:
		return node.updateActions(got)
	}

	return nil, fmt.Errorf("ForwardingRule: invalid plan op %s", op)
}

func (node *ForwardingRuleNode) createActions(got Node) ([]exec.Action, error) {
	return opCreateActions[compute.ForwardingRule, alpha.ForwardingRule, beta.ForwardingRule](&forwardingRuleOps{}, node, node.resource)
}

type forwardingRuleUpdateAction struct {
	exec.ActionBase

	id *cloud.ResourceID
	// target if non-empty will call setTarget(),
	target *cloud.ResourceID
	// oldTarget is the previous target before the update.
	oldTarget *cloud.ResourceID

	// labels if non-nil will call setLabels().
	labels map[string]string
}

func (act *forwardingRuleUpdateAction) Run(ctx context.Context, cl cloud.Cloud) ([]exec.Event, error) {
	// TODO: project routing.
	if act.labels != nil {
		switch act.id.Key.Type() {
		case meta.Global:
			err := cl.GlobalForwardingRules().SetLabels(ctx, act.id.Key, &compute.GlobalSetLabelsRequest{
				LabelFingerprint: "XXX",
				Labels:           act.labels,
			})
			if err != nil {
				return nil, fmt.Errorf("forwardingRuleUpdateAction Run(%s): SetLabels: %w", act.id, err)
			}
		case meta.Regional:
			err := cl.ForwardingRules().SetLabels(ctx, act.id.Key, &compute.RegionSetLabelsRequest{
				LabelFingerprint: "XXX",
				Labels:           act.labels,
			})
			if err != nil {
				return nil, fmt.Errorf("forwardingRuleUpdateAction Run(%s): SetLabels: %w", act.id, err)
			}
		default:
			return nil, fmt.Errorf("forwardingRuleUpdateAction Run(%s): invalid key type", act.id)
		}
	}

	if act.target != nil {
		switch act.id.Key.Type() {
		case meta.Global:
			err := cl.GlobalForwardingRules().SetTarget(ctx, act.id.Key, &compute.TargetReference{
				Target: act.target.SelfLink(meta.VersionGA),
			})
			if err != nil {
				return nil, fmt.Errorf("forwardingRuleUpdateAction Run(%s): SetTarget: %w", act.id, err)
			}
		case meta.Regional:
			err := cl.GlobalForwardingRules().SetTarget(ctx, act.id.Key, &compute.TargetReference{
				Target: act.target.SelfLink(meta.VersionGA),
			})
			if err != nil {
				return nil, fmt.Errorf("forwardingRuleUpdateAction Run(%s): SetTarget: %w", act.id, err)
			}
		default:
			return nil, fmt.Errorf("forwardingRuleUpdateAction Run(%s): invalid key type", act.id)
		}
	}

	var events []exec.Event
	if act.oldTarget != nil {
		events = append(events, exec.NewDropRefEvent(act.id, act.oldTarget))
	}

	return events, nil
}

func (act *forwardingRuleUpdateAction) DryRun() []exec.Event {
	// TODO
	return nil
}
func (act *forwardingRuleUpdateAction) String() string {
	return fmt.Sprintf("ForwardingRuleUpdateAction(%s)", act.id)
}
func (act *forwardingRuleUpdateAction) Summary() string { return "TODO" } // TODO: delete me

func parseForwardingRuleTargetValue(nodeID *cloud.ResourceID, v any) (*cloud.ResourceID, error) {
	val, ok := v.(string)
	if !ok {
		return nil, fmt.Errorf("forwardingRuleNode: updateActions: node %s: bad Target type %T", nodeID, v)
	}
	if val == "" {
		return nil, nil
	}
	target, err := cloud.ParseResourceURL(val)
	if err != nil {
		return nil, fmt.Errorf("forwardingRuleNode: updateActions: node %s: bad Target value %q", nodeID, val)
	}
	return target, nil
}

func parseForwardingRuleLabelsValue(nodeID *cloud.ResourceID, v any) (map[string]string, error) {
	val, ok := v.(map[string]string)
	if !ok {
		return nil, fmt.Errorf("forwardingRuleNode: updateActions: node %s: bad Labels type %T", nodeID, v)
	}
	return val, nil
}

func (node *ForwardingRuleNode) updateActions(got Node) ([]exec.Action, error) {
	details := node.LocalPlan().Details()
	if details == nil {
		return nil, fmt.Errorf("forwardingRuleNode: updateActions: node %s has not been planned", node.ID())
	}

	act := &forwardingRuleUpdateAction{id: node.ID()}

	for _, item := range details.Diff.Items {
		switch {
		case api.Path{}.Pointer().Field("Target").Equal(item.Path):
			oldTarget, err := parseForwardingRuleTargetValue(node.ID(), item.A)
			if err != nil {
				return nil, err
			}
			target, err := parseForwardingRuleTargetValue(node.ID(), item.B)
			if err != nil {
				return nil, err
			}
			act.Want = append(act.Want, exec.NewExistsEvent(target))
			act.oldTarget = oldTarget
			act.target = target
		case api.Path{}.Pointer().Field("Labels").Equal(item.Path):
			val, err := parseForwardingRuleLabelsValue(node.ID(), item.B)
			if err != nil {
				return nil, err
			}
			act.labels = val
		default:
			return nil, fmt.Errorf("forwardingRuleNode: updateActions: node %s: bad Fields type %T", node.ID(), item.B)
		}
	}

	return []exec.Action{
		exec.NewExistsEventOnlyAction(node.ID()),
		act,
	}, nil
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
