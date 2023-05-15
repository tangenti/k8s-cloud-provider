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

package workflow

import (
	"context"
	"fmt"

	"github.com/GoogleCloudPlatform/k8s-cloud-provider/pkg/cloud"
	"github.com/GoogleCloudPlatform/k8s-cloud-provider/pkg/cloud/sync"
	"k8s.io/klog/v2"
)

// PlanLoadBalancer will perform updates to cloud resources wanted in graph.
func PlanLoadBalancer(ctx context.Context, c cloud.Cloud, graph *sync.Graph) error {
	w := lbPlanner{
		cloud: c,
		got:   sync.NewGraph(),
		want:  graph,
	}
	if err := w.plan(ctx); err != nil {
		return err
	}

	// XXX
	for _, node := range w.want.All() {
		klog.Infof("[PLAN] %v: %s", node.ID(), node.LocalPlan())
	}

	return nil
}

type lbPlanner struct {
	cloud cloud.Cloud
	got   *sync.Graph
	want  *sync.Graph
}

func (w *lbPlanner) plan(ctx context.Context) error {
	// Assemble the "got" graph. This will get the current state of any
	// resources and also enumerate any resouces that are currently linked that
	// are not in the "want" graph.
	for _, node := range w.want.All() {
		gotNode := node.NewEmptyNode()
		w.got.AddNode(gotNode)
	}

	// Fetch the current resource graph from Cloud.
	// TODO: resource_prefix, ownership due to prefix etc.
	err := sync.TransitiveClosure(ctx, w.cloud, w.got,
		sync.TransitiveClosureOnGet(func(n sync.Node) error {
			n.SetOwnership(sync.OwnershipManaged)
			return nil
		}),
	)
	if err != nil {
		return err
	}

	// Copy nodes in "got" that aren't in "want" as not existing so the planner
	// will mark them for deletion.
	for _, gotNode := range w.got.All() {
		if w.want.Resource(gotNode.ID()) != nil {
			continue
		}
		wantNode := gotNode.NewEmptyNode()
		wantNode.SetState(sync.NodeDoesNotExist)
		w.want.AddNode(wantNode)
	}

	// Compute the local plan for each resource.
	if err := sync.LocalPlan(w.got, w.want); err != nil {
		return err
	}

	if err := w.propagateRecreates(); err != nil {
		return err
	}

	if err := w.sanityCheck(); err != nil {
		return err
	}

	return nil
}

// propagateRecreates through inbound references. If a resource needs to be
// recreated, this means any references will also be affected transitively.
func (w *lbPlanner) propagateRecreates() error {
	var recreateNodes []sync.Node
	for _, node := range w.want.All() {
		if node.LocalPlan().Op() == sync.OpRecreate {
			recreateNodes = append(recreateNodes, node)
		}
	}

	done := map[cloud.ResourceMapKey]bool{}
	for _, node := range recreateNodes {
		for _, inRefNode := range sync.TransitiveInRefs(w.want, node) {
			if done[inRefNode.MapKey()] {
				continue
			}
			done[inRefNode.MapKey()] = true

			if inRefNode.Ownership() != sync.OwnershipManaged {
				return fmt.Errorf("lb workflow: %v planned for recreate, but inRef %v ownership=%s", node.ID(), inRefNode.ID(), inRefNode.Ownership())
			}

			switch inRefNode.LocalPlan().Op() {
			case sync.OpUpdate:
				fallthrough
			case sync.OpNothing:
				node.LocalPlan().Set(sync.Action{
					Operation: sync.OpRecreate,
					Why:       fmt.Sprintf("Dependency %v is being recreated", node.ID()),
				})
			}
		}
	}
	return nil
}

func (w *lbPlanner) sanityCheck() error {
	w.want.ComputeInRefs()
	for _, node := range w.want.All() {
		switch node.LocalPlan().Op() {
		case sync.OpUnknown:
			return fmt.Errorf("lb workflow: node %v has invalid op %s", node.ID(), node.LocalPlan().Op())
		case sync.OpDelete:
			// If A => B; if B is to be deleted, then A must be deleted.
			for _, refs := range node.InRefs() {
				if inNode := w.want.Resource(refs.To); inNode == nil {
					return fmt.Errorf("lb workflow: inRef from node %v that doesn't exist", refs.From)
				} else if inNode.LocalPlan().Op() != sync.OpDelete {
					return fmt.Errorf("lb workflow: %v to be deleted, but inRef %v is not", node.ID(), inNode.ID())
				}
			}
		}
	}
	return nil
}

/*
func (w *lbPlanner) execute(ctx context.Context) error {
	// TODO
	// Create or update BackendServices, HealthChecks. Each BS can be done in
	// parallel.
	//
	// Create/update the rest of the hierarchy. It is preferred to write this
	// first using a single-threaded code and then add parallelism in stages
	// rather than implementing a full-featured task graph from the outset.
	return nil
}
*/
