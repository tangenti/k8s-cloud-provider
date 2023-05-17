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

package plan

import (
	"context"
	"fmt"

	"github.com/GoogleCloudPlatform/k8s-cloud-provider/pkg/cloud"
	"github.com/GoogleCloudPlatform/k8s-cloud-provider/pkg/cloud/sync"
	"github.com/GoogleCloudPlatform/k8s-cloud-provider/pkg/cloud/sync/algo/localplan"
	"github.com/GoogleCloudPlatform/k8s-cloud-provider/pkg/cloud/sync/algo/traversal"
	"github.com/GoogleCloudPlatform/k8s-cloud-provider/pkg/cloud/sync/algo/trclosure"
	"github.com/GoogleCloudPlatform/k8s-cloud-provider/pkg/cloud/sync/exec"
)

// Do will plan updates to cloud resources wanted in graph. Returns the set of
// Actions needed to sync to "want".
func Do(ctx context.Context, c cloud.Cloud, want *sync.Graph) ([]exec.Action, error) {
	w := planner{
		cloud: c,
		want:  want,
	}
	return w.plan(ctx)
}

const errPrefix = "Plan"

type planner struct {
	cloud cloud.Cloud
	got   *sync.Graph
	want  *sync.Graph
}

func (pl *planner) plan(ctx context.Context) ([]exec.Action, error) {
	// Assemble the "got" graph. This will get the current state of any
	// resources and also enumerate any resouces that are currently linked that
	// are not in the "want" graph.
	pl.got = pl.want.CloneWithEmptyNodes()

	// Fetch the current resource graph from Cloud.
	// TODO: resource_prefix, ownership due to prefix etc.
	err := trclosure.Do(ctx, pl.cloud, pl.got,
		trclosure.OnGetFunc(func(n sync.Node) error {
			n.SetOwnership(sync.OwnershipManaged)
			return nil
		}),
	)
	if err != nil {
		return nil, err
	}

	// Figure out what to do with Nodes in "got" that aren't in "want". These
	// are resources that will no longer by referenced in the updated graph.
	for _, gotNode := range pl.got.All() {
		switch {
		case pl.want.Resource(gotNode.ID()) != nil:
			// Node exists in "want", don't need to do anything.
		case gotNode.Ownership() == sync.OwnershipExternal:
			// TODO: clone the node from the "got" graph for "want" unchanged.
		case gotNode.Ownership() == sync.OwnershipManaged:
			// Nodes that are no longer referenced should be deleted.
			wantNode := gotNode.NewEmptyNode()
			wantNode.SetState(sync.NodeDoesNotExist)
			pl.want.AddNode(wantNode)
		default:
			return nil, fmt.Errorf("%s: node %s has invalid ownership %s", errPrefix, gotNode.ID(), gotNode.Ownership())
		}
	}

	// Compute the local plan for each resource.
	if err := localplan.Do(pl.got, pl.want); err != nil {
		return nil, err
	}

	if err := pl.propagateRecreates(); err != nil {
		return nil, err
	}

	if err := pl.sanityCheck(); err != nil {
		return nil, err
	}

	actions, err := pl.want.Actions(pl.got)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", errPrefix, err)
	}
	return actions, nil
}

// propagateRecreates through inbound references. If a resource needs to be
// recreated, this means any references will also be affected transitively.
func (pl *planner) propagateRecreates() error {
	var recreateNodes []sync.Node
	for _, node := range pl.want.All() {
		if node.LocalPlan().Op() == sync.OpRecreate {
			recreateNodes = append(recreateNodes, node)
		}
	}

	done := map[cloud.ResourceMapKey]bool{}
	for _, node := range recreateNodes {
		for _, inRefNode := range traversal.TransitiveInRefs(pl.want, node) {
			if done[inRefNode.MapKey()] {
				continue
			}
			done[inRefNode.MapKey()] = true

			if inRefNode.Ownership() != sync.OwnershipManaged {
				return fmt.Errorf("%s: %v planned for recreate, but inRef %v ownership=%s", errPrefix, node.ID(), inRefNode.ID(), inRefNode.Ownership())
			}

			switch inRefNode.LocalPlan().Op() {
			case sync.OpCreate, sync.OpRecreate, sync.OpDelete:
				// Resource is already being created or destroy.
			case sync.OpNothing, sync.OpUpdate:
				node.LocalPlan().Set(sync.PlanDetails{
					Operation: sync.OpRecreate,
					Why:       fmt.Sprintf("Dependency %v is being recreated", node.ID()),
				})
			default:
				return fmt.Errorf("%s: inRef %s has invalid op %s, can't propagate recreate", errPrefix, inRefNode.ID(), inRefNode.LocalPlan().Op())
			}
		}
	}
	return nil
}

func (pl *planner) sanityCheck() error {
	pl.want.ComputeInRefs()
	for _, node := range pl.want.All() {
		switch node.LocalPlan().Op() {
		case sync.OpUnknown:
			return fmt.Errorf("%s: node %v has invalid op %s", errPrefix, node.ID(), node.LocalPlan().Op())
		case sync.OpDelete:
			// If A => B; if B is to be deleted, then A must be deleted.
			for _, refs := range node.InRefs() {
				if inNode := pl.want.Resource(refs.To); inNode == nil {
					return fmt.Errorf("%s: inRef from node %v that doesn't exist", errPrefix, refs.From)
				} else if inNode.LocalPlan().Op() != sync.OpDelete {
					return fmt.Errorf("%s: %v to be deleted, but inRef %v is not", errPrefix, node.ID(), inNode.ID())
				}
			}
		}
	}

	return nil
}
