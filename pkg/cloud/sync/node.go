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
	"github.com/GoogleCloudPlatform/k8s-cloud-provider/pkg/cloud/api"
	"github.com/GoogleCloudPlatform/k8s-cloud-provider/pkg/cloud/meta"
	"github.com/GoogleCloudPlatform/k8s-cloud-provider/pkg/cloud/sync/exec"
	"github.com/kr/pretty"
)

type OwnershipStatus string

var (
	OwnershipUnknown  OwnershipStatus = "Unknown"
	OwnershipManaged  OwnershipStatus = "Managed"
	OwnershipExternal OwnershipStatus = "External"
)

// NodeState is the state of the node in the Graph.
type NodeState string

const (
	NodeUnknown      NodeState = "Unknown"
	NodeExists       NodeState = "Exists"
	NodeDoesNotExist NodeState = "DoesNotExist"
	NodeStateError   NodeState = "Error"
)

// ResourceRef identifies a reference from the resource From in the field Path
// to the resource To.
type ResourceRef struct {
	From *cloud.ResourceID
	Path api.Path
	To   *cloud.ResourceID
}

type NodeAttributes struct {
	Ownership OwnershipStatus
}

// Node in the resource graph.
type Node interface {
	// State of the node.
	State() NodeState
	// SetState of the node.
	SetState(NodeState)

	// ID uniquely identifying this resource.
	ID() *cloud.ResourceID
	// MapKey returns a key usable for uniquely identifying in a map.
	MapKey() cloud.ResourceMapKey

	// Ownership of this resource.
	Ownership() OwnershipStatus
	// SetOwnership of this resource.
	SetOwnership(OwnershipStatus)

	// Sprint pretty-prints the Node.
	Sprint() string

	// OutRefs of this resource pointing to other resources.
	OutRefs() ([]ResourceRef, error)
	// InRefs pointing to this resource.
	InRefs() []ResourceRef

	// Get the current copy of the resource from the cloud, if any exists. This
	// will take into account the appropriate version (e.g. beta) if it is
	// needed.
	//
	// This is may be one or more blocking calls to the Cloud service.
	Get(ctx context.Context, gcp cloud.Cloud) error
	// GetErr is the error from the last Get() operation.
	GetErr() error

	// NewEmptyNode returns a node that has the same attributes and underlying
	// type but has no contents in the resource. This is used to populate a
	// graph for getting the current state from Cloud (i.e. the "got" graph).
	NewEmptyNode() Node

	// Diff this node (want) with the state of the Node (got). This computes
	// whether the Sync operation will be an update or recreation.
	Diff(got Node) (*Action, error)

	// LocalPlan returns the plan for updating this Node. This will be nil for
	// graphs that have not been planned.
	LocalPlan() *Plan

	// Actions needed to perform the plan. This will be empty for graphs that
	// have not been planned. "got" is the current state of the Node in the
	// "got" graph.
	Actions(got Node) ([]exec.Action, error)

	// Internal methods.
	clearInRefs()
	addInRef(ref ResourceRef)
	unTypedResource() untypedResource
	setResource(ur untypedResource)
	setAttributes(na NodeAttributes)
	version() meta.Version
	setVersion(meta.Version)
}

// untypedResource is the type-erased version of FrozenResource.
type untypedResource interface {
	ResourceID() *cloud.ResourceID
	Version() meta.Version
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

// nodeBase contains common fields across all of the Graph nodes.
type nodeBase[GA any, Alpha any, Beta any] struct {
	id        *cloud.ResourceID
	typeTrait api.TypeTrait[GA, Alpha, Beta]

	state     NodeState
	ownership OwnershipStatus
	resource  api.FrozenResource[GA, Alpha, Beta]

	inRefs []ResourceRef
	getErr error

	// getVer determines the Version of the resource to get from Cloud.
	getVer meta.Version
	// plan is only set in the "want" graph.
	plan Plan
}

func (n *nodeBase[GA, Alpha, Beta]) Resource() api.FrozenResource[GA, Alpha, Beta] { return n.resource }

func (n *nodeBase[GA, Alpha, Beta]) State() NodeState               { return n.state }
func (n *nodeBase[GA, Alpha, Beta]) SetState(s NodeState)           { n.state = s }
func (n *nodeBase[GA, Alpha, Beta]) ID() *cloud.ResourceID          { return n.id }
func (n *nodeBase[GA, Alpha, Beta]) MapKey() cloud.ResourceMapKey   { return n.id.MapKey() }
func (n *nodeBase[GA, Alpha, Beta]) Ownership() OwnershipStatus     { return n.ownership }
func (n *nodeBase[GA, Alpha, Beta]) SetOwnership(o OwnershipStatus) { n.ownership = o }
func (n *nodeBase[GA, Alpha, Beta]) InRefs() []ResourceRef          { return n.inRefs }
func (n *nodeBase[GA, Alpha, Beta]) GetErr() error                  { return n.getErr }
func (n *nodeBase[GA, Alpha, Beta]) Sprint() string                 { return pretty.Sprint(n) }
func (n *nodeBase[GA, Alpha, Beta]) LocalPlan() *Plan               { return &n.plan }

func (n *nodeBase[GA, Alpha, Beta]) init(id *cloud.ResourceID, tt api.TypeTrait[GA, Alpha, Beta]) {
	n.id = id
	n.typeTrait = tt
	n.state = NodeUnknown
	n.ownership = OwnershipUnknown
	n.getVer = meta.VersionGA
}

func (n *nodeBase[GA, Alpha, Beta]) initEmptyNode(other Node) {
	n.ownership = other.Ownership()
	n.getVer = other.version()
}

func (n *nodeBase[GA, Alpha, Beta]) clearInRefs()                     { n.inRefs = []ResourceRef{} }
func (n *nodeBase[GA, Alpha, Beta]) addInRef(ref ResourceRef)         { n.inRefs = append(n.inRefs, ref) }
func (n *nodeBase[GA, Alpha, Beta]) unTypedResource() untypedResource { return n.resource }
func (n *nodeBase[GA, Alpha, Beta]) setAttributes(na NodeAttributes)  { n.ownership = na.Ownership }
func (n *nodeBase[GA, Alpha, Beta]) version() meta.Version            { return n.getVer }
func (n *nodeBase[GA, Alpha, Beta]) setVersion(v meta.Version)        { n.getVer = v }

func (n *nodeBase[GA, Alpha, Beta]) setResource(ur untypedResource) {
	r, ok := ur.(api.FrozenResource[GA, Alpha, Beta])
	if !ok {
		panic(fmt.Sprintf("setResource: invalid type %T", ur))
	}
	n.resource = r
	n.getVer = r.Version()
}

func createPreconditions(node Node) ([]exec.Event, error) {
	outRefs, err := node.OutRefs()
	if err != nil {
		return nil, err
	}
	var events []exec.Event
	// Condition: references must exist before creation.
	for _, ref := range outRefs {
		events = append(events, exec.NewExistsEvent(ref.To))
	}
	return events, nil
}

func deletePreconditions(node Node) []exec.Event {
	var ret []exec.Event

	// Event: all references are removed. XXX-- this is wrong, these need to come from references in the got graph

	// Condition: resource must exist. TODO: this seems to make issues
	// ret = append(ret, exec.NewExistsEvent(node.ID()))
	// Condition: no inRefs to the resource.
	for _, ref := range node.InRefs() {
		ret = append(ret, exec.NewDropRefEvent(ref.From, ref.To))
	}

	return ret
}

func updatePreconditions(node Node) []exec.Event {
	// Update can only occur if the resource Exists TODO: is there a case where
	// the ambient signal for existance from Update op collides with a
	// reference to it?
	return nil // TODO: finish me
}
