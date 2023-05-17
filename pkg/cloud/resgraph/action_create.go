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
	"context"
	"fmt"

	"github.com/GoogleCloudPlatform/k8s-cloud-provider/pkg/cloud"
	"github.com/GoogleCloudPlatform/k8s-cloud-provider/pkg/cloud/api"
	"github.com/GoogleCloudPlatform/k8s-cloud-provider/pkg/cloud/resgraph/exec"
)

func opCreateActions[GA any, Alpha any, Beta any](
	ops genericOps[GA, Alpha, Beta],
	node Node,
	resource api.FrozenResource[GA, Alpha, Beta],
) ([]exec.Action, error) {
	events, err := createPreconditions(node)
	if err != nil {
		return nil, err
	}
	return []exec.Action{
		newGenericCreateAction(events, ops, node.ID(), resource),
	}, nil
}

func newGenericCreateAction[GA any, Alpha any, Beta any](
	want []exec.Event,
	ops genericOps[GA, Alpha, Beta],
	id *cloud.ResourceID,
	resource api.FrozenResource[GA, Alpha, Beta],
) *genericCreateAction[GA, Alpha, Beta] {
	return &genericCreateAction[GA, Alpha, Beta]{
		ActionBase: exec.ActionBase{Want: want},
		ops:        ops,
		id:         id,
		resource:   resource,
	}
}

type genericCreateAction[GA any, Alpha any, Beta any] struct {
	exec.ActionBase
	ops      genericOps[GA, Alpha, Beta]
	id       *cloud.ResourceID
	resource api.FrozenResource[GA, Alpha, Beta]
}

func (a *genericCreateAction[GA, Alpha, Beta]) Run(
	ctx context.Context,
	c cloud.Cloud,
) ([]exec.Event, error) {
	err := a.ops.createFuncs(c).do(ctx, a.id, a.resource)
	return []exec.Event{exec.NewExistsEvent(a.id)}, err
}

func (a *genericCreateAction[GA, Alpha, Beta]) DryRun() []exec.Event {
	return []exec.Event{exec.NewExistsEvent(a.id)}
}

func (a *genericCreateAction[GA, Alpha, Beta]) String() string {
	return fmt.Sprintf("GenericCreateAction(%v)", a.id)
}

func createPreconditions(want Node) ([]exec.Event, error) {
	outRefs, err := want.OutRefs()
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
