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

package exec

import (
	"context"
	"fmt"

	"github.com/GoogleCloudPlatform/k8s-cloud-provider/pkg/cloud"
)

// Action is an operation that updates external resources.
type Action interface {
	// CanRun returns true if all of the Events this Action is waiting for have
	// been signaled.
	CanRun() bool
	// Signal this Action with the Event that has occurred.
	// TODO: error propagation.
	// Returns true if the Action was waiting on the Event.
	Signal(Event) bool
	// Run the Action, performing the operations. Returns a list of Events to
	// signal and/or any errors that have occurred.
	Run(context.Context, cloud.Cloud) ([]Event, error)
	// DryRun simulates running the Action. Returns a list of Events to signal.
	DryRun() []Event
	// String returns a human-readable representation of the Action for logging.
	String() string
}

// ActionBase is a helper that implements some standard behaviors of common
// Action implementation.
type ActionBase struct {
	// Want are the events this action is still waiting for.
	Want []Event
	// Done tracks the events that have happened. This is for debugging.
	Done []Event
}

func (b *ActionBase) CanRun() bool { return len(b.Want) == 0 }

func (b *ActionBase) Signal(ev Event) bool {
	for i, wantEv := range b.Want {
		if wantEv.Equal(ev) {
			b.Want = append(b.Want[0:i], b.Want[i+1:]...)
			b.Done = append(b.Done, wantEv)

			return true
		}
	}
	return false
}

// NewExistsEventOnlyAction returns an eventOnlyAction. This Action Signals events
// that represent the existance of a resource at the start of execution. This is
// a synthetic event will be signaled at the start of execution before any
// Actions are performed.
func NewExistsEventOnlyAction(id *cloud.ResourceID) Action {
	return &eventOnlyAction{
		events: []Event{&existsEvent{id: id}},
	}
}

// eventOnlyAction exist only to signal events that already happened (e.g.
// resource already exists). These Actions do not have side-effect; they only
// exist to model the starting conditions of an execution.
type eventOnlyAction struct {
	events []Event
}

func (b *eventOnlyAction) CanRun() bool      { return true }
func (b *eventOnlyAction) Signal(Event) bool { return false }
func (a *eventOnlyAction) String() string    { return fmt.Sprintf("EventOnlyAction(%v)", a.events) }
func (a *eventOnlyAction) DryRun() []Event   { return a.events }

func (a *eventOnlyAction) Run(context.Context, cloud.Cloud) ([]Event, error) { return a.events, nil }

func newTestAction(name string, want, events []Event) *testAction {
	return &testAction{
		ActionBase: ActionBase{
			Want: want,
		},
		name:   name,
		events: events,
	}
}

type testAction struct {
	ActionBase
	name   string
	events []Event
}

func (a *testAction) String() string  { return fmt.Sprintf("TestAction(%s,%v)", a.name, a.events) }
func (a *testAction) DryRun() []Event { return a.events }
func (a *testAction) Run(context.Context, cloud.Cloud) ([]Event, error) {
	return a.events, nil
}
