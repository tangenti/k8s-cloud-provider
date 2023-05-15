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
