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

	"github.com/GoogleCloudPlatform/k8s-cloud-provider/pkg/cloud"
	"github.com/GoogleCloudPlatform/k8s-cloud-provider/pkg/cloud/api"
	"github.com/GoogleCloudPlatform/k8s-cloud-provider/pkg/cloud/sync/exec"
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

func opDeleteActions[GA any, Alpha any, Beta any](
	ops genericOps[GA, Alpha, Beta],
	node Node,
) ([]exec.Action, error) {
	return []exec.Action{
		newGenericDeleteAction(deletePreconditions(node), ops, node.ID()),
	}, nil
}

func opRecreateActions[GA any, Alpha any, Beta any](
	ops genericOps[GA, Alpha, Beta],
	node Node,
	resource api.FrozenResource[GA, Alpha, Beta],
) ([]exec.Action, error) {
	deleteAction := newGenericDeleteAction(deletePreconditions(node), ops, node.ID())

	createEvents, err := createPreconditions(node)
	if err != nil {
		return nil, err
	}
	// Condition: resource must have been deleted.
	createEvents = append(createEvents, exec.NewNotExistsEvent(node.ID()))
	createAction := newGenericCreateAction(createEvents, ops, node.ID(), resource)

	return []exec.Action{deleteAction, createAction}, nil
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
	return "GenericCreateAction TODO"
}

func newGenericDeleteAction[GA any, Alpha any, Beta any](
	want []exec.Event,
	ops genericOps[GA, Alpha, Beta],
	id *cloud.ResourceID,
) *genericDeleteAction[GA, Alpha, Beta] {
	return &genericDeleteAction[GA, Alpha, Beta]{
		ActionBase: exec.ActionBase{Want: want},
		ops:        ops,
		id:         id,
	}
}

type genericDeleteAction[GA any, Alpha any, Beta any] struct {
	exec.ActionBase
	ops genericOps[GA, Alpha, Beta]
	id  *cloud.ResourceID
}

func (a *genericDeleteAction[GA, Alpha, Beta]) Run(
	ctx context.Context,
	c cloud.Cloud,
) ([]exec.Event, error) {
	err := a.ops.deleteFuncs(c).do(ctx, a.id)
	return []exec.Event{exec.NewNotExistsEvent(a.id)}, err
}

func (a *genericDeleteAction[GA, Alpha, Beta]) DryRun() []exec.Event {
	return []exec.Event{exec.NewNotExistsEvent(a.id)}
}

func (a *genericDeleteAction[GA, Alpha, Beta]) String() string {
	return "GenericDeleteAction TODO"
}

// TODO
type genericUpdateAction[GA any, Alpha any, Beta any] struct {
	exec.ActionBase
}

func (a *genericUpdateAction[GA, Alpha, Beta]) Run(
	ctx context.Context,
	c cloud.Cloud,
) ([]exec.Event, error) {
	return nil, nil
}

func (a *genericUpdateAction[GA, Alpha, Beta]) DryRun() []exec.Event {
	return nil
}

func (a *genericUpdateAction[GA, Alpha, Beta]) String() string {
	return "GenericUpdateAction TODO"
}
