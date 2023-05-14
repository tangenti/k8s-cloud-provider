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
			&existsEvent{id: id},
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
