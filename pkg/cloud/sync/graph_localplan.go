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
	"fmt"
)

// LocalPlan computes a plan local to each Node in the graph and puts the
// resulting plan in the "want" Graph. It is required that got and want have the
// same set of Nodes; Nodes that don't exist need to be marked as with
// NodeStateDoesNotExist.
func LocalPlan(got, want *Graph) error {
	p := localPlanner{got: got, want: want}
	return p.do()
}

type localPlanner struct {
	got  *Graph
	want *Graph
}

func (p *localPlanner) do() error {
	if err := p.preconditions(); err != nil {
		return err
	}
	for _, gotNode := range p.got.All() {
		wantNode := p.want.Resource(gotNode.ID())
		// Preconditions check that wantNode is not nil.
		if err := p.localPlan(gotNode, wantNode); err != nil {
			return err
		}
	}

	return nil
}

func (p *localPlanner) preconditions() error {
	for _, node := range p.got.All() {
		if p.want.Resource(node.ID()) == nil {
			return fmt.Errorf("localPlanner.preconditions: node %s in got but not in want", node.ID())
		}
	}
	for _, node := range p.want.All() {
		if p.got.Resource(node.ID()) == nil {
			return fmt.Errorf("localPlanner.preconditions: node %s in want but not in got", node.ID())
		}
	}

	return nil
}

func (p *localPlanner) localPlan(gotNode, wantNode Node) error {
	if wantNode.Ownership() != OwnershipManaged {
		wantNode.LocalPlan().Set(PlanAction{
			Operation: OpNothing,
			Why:       "Node is not managed",
		})
		return nil
	}

	type s struct{ got, want NodeState }

	statePair := s{gotNode.State(), wantNode.State()}
	switch statePair {
	case s{NodeExists, NodeExists}:
		action, err := wantNode.Diff(gotNode)
		if err != nil {
			return fmt.Errorf("localPlanner: %w", err)
		}
		wantNode.LocalPlan().Set(*action)

	case s{NodeExists, NodeDoesNotExist}:
		wantNode.LocalPlan().Set(PlanAction{
			Operation: OpDelete,
			Why:       "Node doesn't exist in want, but exists in got",
		})

	case s{NodeDoesNotExist, NodeExists}:
		wantNode.LocalPlan().Set(PlanAction{
			Operation: OpCreate,
			Why:       "Node doesn't exist in got, but exists in want",
		})

	case s{NodeDoesNotExist, NodeDoesNotExist}:
		wantNode.LocalPlan().Set(PlanAction{
			Operation: OpNothing,
			Why:       "Node does not exist",
		})

	default:
		return fmt.Errorf("nodes are in an invalid state for planning: %+v", statePair)
	}

	return nil
}
