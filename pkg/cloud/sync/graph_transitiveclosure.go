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
	"github.com/kr/pretty"
	"k8s.io/klog/v2"
)

type nodeQueue struct{ items []Node }

func (q *nodeQueue) pop() Node {
	node := q.items[0]
	q.items = q.items[1:]
	return node
}
func (q *nodeQueue) add(n Node)  { q.items = append(q.items, n) }
func (q *nodeQueue) empty() bool { return len(q.items) == 0 }

type TransitiveClosureOption func(c *transitiveClosureConfig)

func TransitiveClosureOnGet(f func(n Node) error) TransitiveClosureOption {
	return func(c *transitiveClosureConfig) { c.onGet = f }
}

type transitiveClosureConfig struct {
	onGet func(n Node) error
}

// TransitiveClosure traverses and fetches the the graph, adding all of the
// dependencies into the graph. onGet() is called after fetching the resource.
// This can modify properties of the Node, for example, set the appropriate
// Onwership state.
func TransitiveClosure(ctx context.Context, cl cloud.Cloud, gr *Graph, opts ...TransitiveClosureOption) error {
	klog.Infof("TransitiveClosure")

	config := transitiveClosureConfig{
		onGet: func(Node) error { return nil },
	}
	for _, o := range opts {
		o(&config)
	}

	// TODO: run this in parallel.
	work := nodeQueue{items: gr.All()}

	for !work.empty() {
		node := work.pop()

		err := node.Get(ctx, cl)
		klog.Infof("node.Get() = %v (%s)", err, pretty.Sprint(node))

		if err != nil {
			return fmt.Errorf("TransitiveClosure: %w", err)
		}
		err = config.onGet(node)
		if err != nil {
			return fmt.Errorf("TransitiveClosure: %w", err)
		}

		if node.Ownership() == OwnershipExternal {
			klog.Infof("%s externally owned", node.ID())
			continue
		}

		outRefs, err := node.OutRefs()
		if err != nil {
			return fmt.Errorf("TransitiveClosure: %w", err)
		}

		for _, ref := range outRefs {
			if gr.Resource(ref.To) != nil {
				continue
			}
			refNode, err := NewNodeByID(ref.To, NodeAttributes{})
			if err != nil {
				return fmt.Errorf("TransitiveClosure: %w", err)
			}
			err = gr.AddNode(refNode)
			if err != nil {
				return fmt.Errorf("TransitiveClosure: %w", err)
			}
			work.add(refNode)
		}
	}

	return nil
}
