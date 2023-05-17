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

package trclosure

import (
	"context"
	"fmt"

	"github.com/GoogleCloudPlatform/k8s-cloud-provider/pkg/cloud"
	"github.com/GoogleCloudPlatform/k8s-cloud-provider/pkg/cloud/sync"
	"github.com/GoogleCloudPlatform/k8s-cloud-provider/pkg/cloud/sync/algo"
	"github.com/kr/pretty"
	"k8s.io/klog/v2"
)

type Option func(c *Config)

// OnGetFunc is called on the Node after getting the resource from Cloud. This
// can modify properties of the Node, for example, set the appropriate Onwership
// state.
func OnGetFunc(f func(n sync.Node) error) Option {
	return func(c *Config) { c.onGet = f }
}

// Config for the algorithm.
type Config struct {
	onGet func(n sync.Node) error
}

// Do traverses and fetches the the graph, adding all of the dependencies into
// the graph, pulling the resource from Cloud as needed.
func Do(ctx context.Context, cl cloud.Cloud, gr *sync.Graph, opts ...Option) error {
	klog.Infof("TransitiveClosure")

	config := Config{
		onGet: func(sync.Node) error { return nil },
	}
	for _, o := range opts {
		o(&config)
	}

	// TODO: run this in parallel.
	work := algo.NewNodeQueue(gr.All())

	for !work.Empty() {
		node := work.Pop()

		err := node.Get(ctx, cl)
		klog.Infof("node.Get() = %v (%s)", err, pretty.Sprint(node))

		if err != nil {
			return fmt.Errorf("TransitiveClosure: %w", err)
		}
		err = config.onGet(node)
		if err != nil {
			return fmt.Errorf("TransitiveClosure: %w", err)
		}

		if node.Ownership() == sync.OwnershipExternal {
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
			refNode, err := sync.NewNodeByID(ref.To, sync.NodeAttributes{})
			if err != nil {
				return fmt.Errorf("TransitiveClosure: %w", err)
			}
			err = gr.AddNode(refNode)
			if err != nil {
				return fmt.Errorf("TransitiveClosure: %w", err)
			}
			work.Add(refNode)
		}
	}

	return nil
}
