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
		a.Signal(ev)
	}
}
