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

package gc

import (
	"context"
	"fmt"
	"strings"

	"github.com/GoogleCloudPlatform/k8s-cloud-provider/pkg/cloud"
	"github.com/GoogleCloudPlatform/k8s-cloud-provider/pkg/cloud/sync"
	"google.golang.org/api/compute/v1"
	"k8s.io/klog/v2"
)

type Scope string

const (
	ScopeGlobal   Scope = "ScopeGlobal"
	ScopeRegional Scope = "ScopeRegional"
)

type Config struct {
	// Scope of the garbage collection operation.
	Scope Scope
	// Region to operate on if ScopeRegional.
	Region string
	// ResourcePrefix is the prefix (e.g. "gke-gw") used to filter
	// the resources that are relevant for the garbage collection
	// operation.
	ResourcePrefix string
	// NodeIsLive returns true if the Node should not be deleted.
	NodeIsLive func(sync.Node) bool
}

func GarbageCollect(ctx context.Context, cl cloud.Cloud, config *Config) error {
	w := workflow{cloud: cl, config: config}
	return w.do(ctx)
}

type workflow struct {
	cloud  cloud.Cloud
	config *Config
}

func (w *workflow) do(ctx context.Context) error {
	var (
		frules []*compute.ForwardingRule
		err    error
	)

	switch w.config.Scope {
	case ScopeGlobal:
		frules, err = w.cloud.GlobalForwardingRules().List(ctx, nil)
		if err != nil {
			return fmt.Errorf("gc workflow: %w", err)
		}
	case ScopeRegional:
		frules, err = w.cloud.ForwardingRules().List(ctx, w.config.Region, nil)
		if err != nil {
			return fmt.Errorf("gc workflow: %w", err)
		}
	default:
		return fmt.Errorf("gc workflow: invalid scope %q", w.config.Scope)
	}

	// Seed the graph with the forwarding rules.
	got := sync.NewGraph()
	for _, fr := range frules {
		if !strings.HasPrefix(fr.Name, w.config.ResourcePrefix) {
			continue
		}
		id, err := cloud.ParseResourceURL(fr.SelfLink)
		if err != nil {
			return fmt.Errorf("gc workflow: %w", err)
		}
		node, err := sync.NewNodeByID(id, sync.NodeAttributes{
			Ownership: sync.OwnershipManaged,
		})
		if err := got.AddNode(node); err != nil {
			return fmt.Errorf("gc workflow: %w", err)
		}
	}
	// Take the transitive closure from the FRs to get the resource graph.
	if err := sync.TransitiveClosure(ctx, w.cloud, got,
		sync.TransitiveClosureOnGet(func(node sync.Node) error {
			if strings.HasPrefix(node.ID().Key.Name, w.config.ResourcePrefix) {
				node.SetOwnership(sync.OwnershipManaged)
			} else {
				node.SetOwnership(sync.OwnershipExternal)
			}
			return nil
		}),
	); err != nil {
		return fmt.Errorf("gc workflow: %w", err)
	}

	// TODO: Make subgraphs live.

	want := sync.NewGraph()
	for _, node := range got.All() {
		if w.config.NodeIsLive(node) {
			// TODO
		} else {
			// Mark this node for deletion in the want graph.
			wantNode := node.NewEmptyNode()
			wantNode.SetState(sync.NodeDoesNotExist)
			if err := want.AddNode(wantNode); err != nil {
				return fmt.Errorf("gc workflow: %w", err)
			}
		}
	}

	klog.Infof("GC want = %+v", want)

	// TODO:

	if err := sync.LocalPlan(got, want); err != nil {
		return fmt.Errorf("gc worflow: %w", err)
	}

	for _, node := range want.All() {
		klog.Infof("[PLAN] node %s: %s", node.ID(), node.LocalPlan())
	}

	// TODO

	return nil
}