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

// Event is triggered during execution, signalling that dependent Actions can
// become possible to run.
type Event interface {
	Equal(other Event) bool
	String() string
	// TODO: handle errors
	// MapKey() string for a more efficient implementation
}

type DropRefEvent struct {
	From, To *cloud.ResourceID
}

func (e *DropRefEvent) Equal(other Event) bool {
	switch other := other.(type) {
	case *DropRefEvent:
		return e.From.Equal(other.From) && e.To.Equal(other.To)
	}
	return false
}

func (e *DropRefEvent) String() string {
	return fmt.Sprintf("DropRef(%v => %v)\n", e.From, e.To)
}

func NewExistsEvent(id *cloud.ResourceID) Event { return &ExistsEvent{ID: id} }

type ExistsEvent struct{ ID *cloud.ResourceID }

func (e *ExistsEvent) Equal(other Event) bool {
	switch other := other.(type) {
	case *ExistsEvent:
		return e.ID.Equal(other.ID)
	}
	return false
}

func (e *ExistsEvent) String() string { return fmt.Sprintf("Exists(%v)", e.ID) }

func NewNotExistsEvent(id *cloud.ResourceID) Event { return &NotExistsEvent{ID: id} }

type NotExistsEvent struct{ ID *cloud.ResourceID }

func (e *NotExistsEvent) Equal(other Event) bool {
	switch other := other.(type) {
	case *NotExistsEvent:
		return e.ID.Equal(other.ID)
	}
	return false
}

func (e *NotExistsEvent) String() string { return fmt.Sprintf("NotExists(%v)", e.ID) }

type Action interface {
	CanRun() bool
	Update(Event) // TODO error

	Run(context.Context, cloud.Cloud) ([]Event, error)
	String() string
}

type ActionBase struct {
	// Want are the events this action is still waiting for.
	Want []Event
	// Done tracks the events that have happened. This is for debugging.
	Done []Event
}

func (b *ActionBase) CanRun() bool { return len(b.Want) == 0 }

func (b *ActionBase) Update(ev Event) {
	for i, wantEv := range b.Want {
		if wantEv.Equal(ev) {
			b.Want = append(b.Want[0:i], b.Want[i+1:]...)
			b.Done = append(b.Done, wantEv)
		}
	}
}

func NewExistsEventAction(id *cloud.ResourceID) Action {
	return &eventOnlyAction{
		events: []Event{
			&ExistsEvent{ID: id},
		},
	}
}

// eventOnlyAction exist only to signal events that already happened (e.g.
// resource already exists).
type eventOnlyAction struct {
	events []Event
}

func (b *eventOnlyAction) CanRun() bool   { return true }
func (b *eventOnlyAction) Update(Event)   {}
func (a *eventOnlyAction) String() string { return fmt.Sprintf("EventOnlyAction(%v)", a.events) }

func (a *eventOnlyAction) Run(context.Context, cloud.Cloud) ([]Event, error) { return a.events, nil }

type Executor interface {
	// Run the actions. Returns the remaining actions that could not be
	// completed and which errors occurred during execution.
	Run(context.Context, cloud.Cloud) ([]Action, error)
}

func NewExecutor(initialEvents []Event, pending []Action) Executor {
	return &serialExecutor{
		pending: pending,
	}
}

type serialExecutor struct {
	pending []Action
	done    []Action
}

func (ex *serialExecutor) Run(ctx context.Context, c cloud.Cloud) ([]Action, error) {
	ex.runEventOnly(ctx, c)

	for a := ex.next(); a != nil; a = ex.next() {
		fmt.Println("executor loop")

		events, err := a.Run(ctx, c)
		if err != nil {
			// TODO: maybe handle propagating the error?
			return ex.pending, err
		}
		ex.done = append(ex.done, a)
		for _, ev := range events {
			fmt.Printf("signal %s\n", ev)
			ex.signal(ev)
		}
	}

	return ex.pending, nil
}

func (ex *serialExecutor) runEventOnly(ctx context.Context, c cloud.Cloud) {
	var nonEventOnly []Action
	for _, a := range ex.pending {
		switch a := a.(type) {
		case *eventOnlyAction:
			evs, _ := a.Run(ctx, c)
			for _, ev := range evs {
				ex.signal(ev)
			}
		default:
			nonEventOnly = append(nonEventOnly, a)
		}
	}
	ex.pending = nonEventOnly
}

func (ex *serialExecutor) next() Action {
	for i, a := range ex.pending {
		if a.CanRun() {
			ex.pending = append(ex.pending[0:i], ex.pending[i+1:]...)
			return a
		}
	}
	return nil
}

func (ex *serialExecutor) signal(ev Event) {
	for _, a := range ex.pending {
		a.Update(ev)
	}
}
