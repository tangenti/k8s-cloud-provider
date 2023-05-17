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

package algo

import "github.com/GoogleCloudPlatform/k8s-cloud-provider/pkg/cloud/sync"

func NewNodeQueue(nodes []sync.Node) *NodeQueue { return &NodeQueue{items: nodes} }

type NodeQueue struct{ items []sync.Node }

func (q *NodeQueue) Pop() sync.Node {
	node := q.items[0]
	q.items = q.items[1:]
	return node
}
func (q *NodeQueue) Add(n sync.Node) { q.items = append(q.items, n) }
func (q *NodeQueue) Empty() bool     { return len(q.items) == 0 }
