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
	"github.com/GoogleCloudPlatform/k8s-cloud-provider/pkg/cloud"
)

// ConnectedSubgraph returns the set of Nodes (inclusive of node) that are
// connected by inbound (InRef) and outbound (OutRef) references.
func ConnectedSubgraph(g *Graph, node Node) ([]Node, error) {
	done := map[cloud.ResourceMapKey]Node{}

	var work nodeQueue
	work.add(node)

	for !work.empty() {
		cur := work.pop()
		done[cur.MapKey()] = cur

		refs, err := cur.OutRefs()
		if err != nil {
			return nil, err
		}
		for _, ref := range refs {
			if _, ok := done[ref.To.MapKey()]; ok {
				continue
			}
			work.add(g.Resource(ref.To))
		}
		refs = cur.InRefs()
		for _, ref := range refs {
			if _, ok := done[ref.From.MapKey()]; ok {
				continue
			}
			work.add(g.Resource(ref.From))
		}
	}

	var ret []Node
	for _, node := range done {
		ret = append(ret, node)
	}

	return ret, nil
}

// TransitiveInRefs returns the set of Nodes (inclusive of node) that point into
// the node. For example, for graph A => B => C; D => B, this will return
// [A, B, D] for B.
func TransitiveInRefs(g *Graph, node Node) []Node {
	g.ComputeInRefs()

	done := map[cloud.ResourceMapKey]Node{}

	var work nodeQueue
	work.add(node)

	for !work.empty() {
		cur := work.pop()
		done[cur.MapKey()] = cur

		refs := cur.InRefs()
		for _, ref := range refs {
			if _, ok := done[ref.From.MapKey()]; ok {
				continue
			}
			work.add(g.Resource(ref.From))
		}
	}

	var ret []Node
	for _, node := range done {
		ret = append(ret, node)
	}

	return ret
}