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

	"github.com/GoogleCloudPlatform/k8s-cloud-provider/pkg/cloud"
	"github.com/GoogleCloudPlatform/k8s-cloud-provider/pkg/cloud/resgraph/exec"
)

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

func updatePreconditions(got, want Node) []exec.Event {
	// Update can only occur if the resource Exists TODO: is there a case where
	// the ambient signal for existance from Update op collides with a
	// reference to it?
	return nil // TODO: finish me
}
