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
	"github.com/GoogleCloudPlatform/k8s-cloud-provider/pkg/cloud/sync/exec"
)

func newGenericDeleteAction[GA any, Alpha any, Beta any](
	want []exec.Event,
	ops genericOps[GA, Alpha, Beta],
	got Node,
) *genericDeleteAction[GA, Alpha, Beta] {
	// Error

	outRefs, _ := got.OutRefs()

	return &genericDeleteAction[GA, Alpha, Beta]{
		ActionBase: exec.ActionBase{Want: want},
		ops:        ops,
		id:         got.ID(),
		outRefs:    outRefs,
	}
}

func opDeleteActions[GA any, Alpha any, Beta any](
	ops genericOps[GA, Alpha, Beta],
	got, want Node,
) ([]exec.Action, error) {
	return []exec.Action{
		newGenericDeleteAction(deletePreconditions(got, want), ops, got),
	}, nil
}

type genericDeleteAction[GA any, Alpha any, Beta any] struct {
	exec.ActionBase
	ops     genericOps[GA, Alpha, Beta]
	id      *cloud.ResourceID
	outRefs []ResourceRef
}

func (a *genericDeleteAction[GA, Alpha, Beta]) Run(
	ctx context.Context,
	c cloud.Cloud,
) ([]exec.Event, error) {
	err := a.ops.deleteFuncs(c).do(ctx, a.id)

	var events []exec.Event
	// Event: Node no longer exists.
	events = append(events, exec.NewNotExistsEvent(a.id))
	for _, ref := range a.outRefs {
		events = append(events, exec.NewDropRefEvent(ref.From, ref.To))
	}

	return events, err
}

func (a *genericDeleteAction[GA, Alpha, Beta]) DryRun() []exec.Event {
	return []exec.Event{exec.NewNotExistsEvent(a.id)}
}

func (a *genericDeleteAction[GA, Alpha, Beta]) String() string {
	return fmt.Sprintf("GenericDeleteAction(%v)", a.id)
}

func deletePreconditions(got, want Node) []exec.Event {
	var ret []exec.Event
	// Condition: no inRefs to the resource still exist.
	for _, ref := range got.InRefs() {
		ret = append(ret, exec.NewDropRefEvent(ref.From, ref.To))
	}
	return ret
}
