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

package traversal

import (
	"github.com/GoogleCloudPlatform/k8s-cloud-provider/pkg/cloud"
	"github.com/GoogleCloudPlatform/k8s-cloud-provider/pkg/cloud/resgraph"
	"github.com/GoogleCloudPlatform/k8s-cloud-provider/pkg/cloud/resgraph/algo"
)

// ConnectedSubgraph returns the set of Nodes (inclusive of node) that are
// connected by inbound (InRef) and outbound (OutRef) references.
func ConnectedSubgraph(g *resgraph.Graph, node resgraph.Node) ([]resgraph.Node, error) {
	done := map[cloud.ResourceMapKey]resgraph.Node{}

	var work algo.NodeQueue
	work.Add(node)

	for !work.Empty() {
		cur := work.Pop()
		done[cur.MapKey()] = cur

		refs, err := cur.OutRefs()
		if err != nil {
			return nil, err
		}
		for _, ref := range refs {
			if _, ok := done[ref.To.MapKey()]; ok {
				continue
			}
			work.Add(g.Resource(ref.To))
		}
		refs = cur.InRefs()
		for _, ref := range refs {
			if _, ok := done[ref.From.MapKey()]; ok {
				continue
			}
			work.Add(g.Resource(ref.From))
		}
	}

	var ret []resgraph.Node
	for _, node := range done {
		ret = append(ret, node)
	}

	return ret, nil
}

// TransitiveInRefs returns the set of Nodes (inclusive of node) that point into
// the node. For example, for graph A => B => C; D => B, this will return
// [A, B, D] for B.
func TransitiveInRefs(g *resgraph.Graph, node resgraph.Node) []resgraph.Node {
	g.ComputeInRefs()

	done := map[cloud.ResourceMapKey]resgraph.Node{}

	var work algo.NodeQueue
	work.Add(node)

	for !work.Empty() {
		cur := work.Pop()
		done[cur.MapKey()] = cur

		refs := cur.InRefs()
		for _, ref := range refs {
			if _, ok := done[ref.From.MapKey()]; ok {
				continue
			}
			work.Add(g.Resource(ref.From))
		}
	}

	var ret []resgraph.Node
	for _, node := range done {
		ret = append(ret, node)
	}

	return ret
}