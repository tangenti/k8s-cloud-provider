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
	"fmt"

	"github.com/GoogleCloudPlatform/k8s-cloud-provider/pkg/cloud"
	"github.com/GoogleCloudPlatform/k8s-cloud-provider/pkg/cloud/resgraph/exec"
)

const graphErrPrefix = "Graph"

// New returns a new empty resource graph.
func New() *Graph {
	ret := &Graph{
		all:                  map[cloud.ResourceMapKey]Node{},
		address:              map[cloud.ResourceMapKey]*AddressNode{},
		backendService:       map[cloud.ResourceMapKey]*BackendServiceNode{},
		fake:                 map[cloud.ResourceMapKey]*fakeNode{},
		forwardingRule:       map[cloud.ResourceMapKey]*ForwardingRuleNode{},
		healthCheck:          map[cloud.ResourceMapKey]*HealthCheckNode{},
		networkEndpointGroup: map[cloud.ResourceMapKey]*NetworkEndpointGroupNode{},
		targetHttpProxy:      map[cloud.ResourceMapKey]*TargetHttpProxyNode{},
		urlMap:               map[cloud.ResourceMapKey]*UrlMapNode{},
	}
	return ret
}

// Graph is a resource graph of GCE resources.
type Graph struct {
	all map[cloud.ResourceMapKey]Node

	address              map[cloud.ResourceMapKey]*AddressNode
	backendService       map[cloud.ResourceMapKey]*BackendServiceNode
	fake                 map[cloud.ResourceMapKey]*fakeNode
	forwardingRule       map[cloud.ResourceMapKey]*ForwardingRuleNode
	healthCheck          map[cloud.ResourceMapKey]*HealthCheckNode
	networkEndpointGroup map[cloud.ResourceMapKey]*NetworkEndpointGroupNode
	targetHttpProxy      map[cloud.ResourceMapKey]*TargetHttpProxyNode
	urlMap               map[cloud.ResourceMapKey]*UrlMapNode
}

// newNodeFromResource is a helper method to create a new node of the right
// specific type.
func newNodeFromResource[N Node, R untypedResource](
	newNode func(newFunc *cloud.ResourceID) N,
	resource R,
	na NodeAttributes,
) N {
	node := newNode(resource.ResourceID())
	node.setResource(resource)
	node.setAttributes(na)
	node.setVersion(resource.Version())

	return node
}

func (g *Graph) AddAddress(res Address, na NodeAttributes) *AddressNode {
	node := newNodeFromResource(newAddressNode, res, na)
	g.AddNode(node)
	return node
}

func (g *Graph) AddBackendService(res BackendService, na NodeAttributes) *BackendServiceNode {
	node := newNodeFromResource(newBackendServiceNode, res, na)
	g.AddNode(node)
	return node
}

func (g *Graph) AddForwardingRule(res ForwardingRule, na NodeAttributes) *ForwardingRuleNode {
	node := newNodeFromResource(newForwardingRuleNode, res, na)
	g.AddNode(node)
	return node
}

func (g *Graph) AddHealthCheck(res HealthCheck, na NodeAttributes) *HealthCheckNode {
	node := newNodeFromResource(newHealthCheckNode, res, na)
	g.AddNode(node)
	return node
}

func (g *Graph) AddNetworkEndpointGroup(res NetworkEndpointGroup, na NodeAttributes) *NetworkEndpointGroupNode {
	node := newNodeFromResource(newNetworkEndpointGroupNode, res, na)
	g.AddNode(node)
	return node
}

func (g *Graph) AddTargetHttpProxy(res TargetHttpProxy, na NodeAttributes) *TargetHttpProxyNode {
	node := newNodeFromResource(newTargetHttpProxyNode, res, na)
	g.AddNode(node)
	return node
}

func (g *Graph) AddUrlMap(res UrlMap, na NodeAttributes) *UrlMapNode {
	node := newNodeFromResource(newUrlMapNode, res, na)
	g.AddNode(node)
	return node
}

func (g *Graph) addFake(res fake, na NodeAttributes) *fakeNode {
	node := newNodeFromResource(newFakeNode, res, na)
	g.AddNode(node)
	return node
}

// Accessors for the resources.
func (g *Graph) Address(id *cloud.ResourceID) *AddressNode { return g.address[id.MapKey()] }
func (g *Graph) BackendService(id *cloud.ResourceID) *BackendServiceNode {
	return g.backendService[id.MapKey()]
}
func (g *Graph) ForwardingRule(id *cloud.ResourceID) *ForwardingRuleNode {
	return g.forwardingRule[id.MapKey()]
}
func (g *Graph) HealthCheck(id *cloud.ResourceID) *HealthCheckNode { return g.healthCheck[id.MapKey()] }
func (g *Graph) NetworkEndpointGroup(id *cloud.ResourceID) *NetworkEndpointGroupNode {
	return g.networkEndpointGroup[id.MapKey()]
}
func (g *Graph) TargetHttpProxy(id *cloud.ResourceID) *TargetHttpProxyNode {
	return g.targetHttpProxy[id.MapKey()]
}
func (g *Graph) UrlMap(id *cloud.ResourceID) *UrlMapNode { return g.urlMap[id.MapKey()] }

// AddNode adds a node to the graph.
func (g *Graph) AddNode(node Node) error {
	key := node.MapKey()
	g.all[key] = node

	switch node := node.(type) {
	case *AddressNode:
		g.address[key] = node
		return nil
	case *BackendServiceNode:
		g.backendService[key] = node
		return nil
	case *ForwardingRuleNode:
		g.forwardingRule[key] = node
		return nil
	case *HealthCheckNode:
		g.healthCheck[key] = node
		return nil
	case *NetworkEndpointGroupNode:
		g.networkEndpointGroup[key] = node
		return nil
	case *TargetHttpProxyNode:
		g.targetHttpProxy[key] = node
		return nil
	case *UrlMapNode:
		g.urlMap[key] = node
		return nil
	}

	return fmt.Errorf("unknown resource type: %T", node)
}

// CloneWithEmptyNodes creates a clone of this graph with Nodes that have the
// same attributes but does not have any values (e.g. the Resource is nil). This
// is used with TransitiveClosure to download the Resource values from Cloud.
func (g *Graph) CloneWithEmptyNodes() *Graph {
	ret := New()
	for _, node := range g.All() {
		ret.AddNode(node.NewEmptyNode())
	}
	return ret
}

// All returns all nodes in the graph.
func (g *Graph) All() []Node {
	var ret []Node
	for _, n := range g.all {
		ret = append(ret, n)
	}
	return ret
}

// Resource returns the resource named by id or nil if not in the graph.
func (g *Graph) Resource(id *cloud.ResourceID) Node { return g.all[id.MapKey()] }

// Actions returns all of the actions for the Nodes in the graph. This is only
// valid for a planned "want" graph. "got" is the graph of the prior resource
// state.
func (g *Graph) Actions(got *Graph) ([]exec.Action, error) {
	var ret []exec.Action
	for _, node := range g.All() {
		actions, err := node.Actions(got.Resource(node.ID()))
		if err != nil {
			return nil, err
		}
		ret = append(ret, actions...)
	}
	return ret, nil
}

// ComputeInRefs must be called after adding nodes to the graph to have a valid
// InRef for resources.
//
// TODO: model this as a different Type to avoid misusage
func (g *Graph) ComputeInRefs() error {
	for _, node := range g.All() {
		node.clearInRefs()
	}
	for _, fromNode := range g.All() {
		refs, err := fromNode.OutRefs()
		if err != nil {
			return err
		}
		for _, ref := range refs {
			toNode, ok := g.all[ref.To.MapKey()]
			if !ok {
				return fmt.Errorf("%s: missing outRef: %s points to %s which isn't in the graph", graphErrPrefix, fromNode.ID(), toNode.ID())
			}
			toNode.addInRef(ref)
		}
	}
	return nil
}

// Validate the graph.
func (g *Graph) Validate() error {
	for _, n := range g.all {
		// No nodes have OwnershipUnknown
		if n.Ownership() == OwnershipUnknown {
			return fmt.Errorf("%s: node %s has ownership %s", graphErrPrefix, n.ID(), n.Ownership())
		}
		// ResourceID is not mismatched
		if !n.unTypedResource().ResourceID().Equal(n.ID()) {
			return fmt.Errorf("%s: node and resource id mismatch (node=%v, id=%v)", graphErrPrefix, n.ID(), n.unTypedResource().ResourceID())
		}
	}
	// All resources have their dependencies in the graph if they are OwnershipManaged.
	for _, n := range g.all {
		if n.Ownership() != OwnershipManaged {
			continue
		}
		deps, err := n.OutRefs()
		if err != nil {
			return err
		}
		for _, d := range deps {
			if _, ok := g.all[d.To.MapKey()]; !ok {
				return fmt.Errorf("%s: missing outRef: %v points to %v which isn't in the graph", graphErrPrefix, n.ID(), d.To)
			}
		}
	}

	return nil
}