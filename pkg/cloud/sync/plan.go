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
	"github.com/GoogleCloudPlatform/k8s-cloud-provider/pkg/cloud/api"
	"github.com/kr/pretty"
)

// Plan for what will be done to the Node.
type Plan struct {
	operation Operation
	// action is a history of Actions that were planned, including previous
	// values Set().
	action []Action
}

// Op to perform.
func (p *Plan) Op() Operation { return p.operation }

// Set the plan to the specified action.
func (p *Plan) Set(a Action) {
	p.operation = a.Operation
	// Save the pervious actions for debugging.
	p.action = append(p.action, a)
}

func (p *Plan) String() string { return pretty.Sprint(p.action) }

// Operation to perform on the Node.
type Operation string

var (
	// OpUnknown means no planning has been done.
	OpUnknown Operation = "Unknown"
	// OpNothing means nothing will happen.
	OpNothing Operation = "Nothing"
	// OpCreate will create the resource.
	OpCreate Operation = "Create"
	// OpRecreate will update the resource by first deleting, then creating
	// the resource. This is necessary as some resources cannot be updated
	// directly in place.
	OpRecreate Operation = "Recreate"
	// OpUpdate will call one or more updates ethods. The specific RPCs called
	// to do the update will be specific to the resource type itself.
	OpUpdate Operation = "Update"
	// OpDelete will delete the resource.
	OpDelete Operation = "Delete"
)

// Action is a human-readable reasons describing the Sync
// operation that has been planned.
type Action struct {
	// Operation associated with this explanation.
	Operation Operation
	// Why is a human readable string describing why this
	// operation was selected.
	Why string
	// Diff is an optional description of the diff between the
	// current and wanted resources.
	Diff *api.DiffResult
}
