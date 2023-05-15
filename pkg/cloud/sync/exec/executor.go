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

	"github.com/GoogleCloudPlatform/k8s-cloud-provider/pkg/cloud"
)

// Executor peforms the operations given by a list of Actions.
type Executor interface {
	// Run the actions. Returns the remaining actions that could not be
	// completed and which errors occurred during execution.
	Run(context.Context, cloud.Cloud) ([]Action, error)
}

type ExecutorOption func(*ExecutorConfig)

func TracerOption(t Tracer) ExecutorOption {
	return func(c *ExecutorConfig) { c.Tracer = t }
}

func DryRunOption(dryRun bool) ExecutorOption {
	return func(c *ExecutorConfig) { c.DryRun = dryRun }
}

type ExecutorConfig struct {
	Tracer Tracer
	DryRun bool
}
