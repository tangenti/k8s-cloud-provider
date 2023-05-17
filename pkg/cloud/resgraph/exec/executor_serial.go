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
	"bytes"
	"context"

	"github.com/GoogleCloudPlatform/k8s-cloud-provider/pkg/cloud"
)

func NewSerialExecutor(pending []Action, opts ...ExecutorOption) *serialExecutor {
	ret := &serialExecutor{
		pending: pending,
	}
	for _, opt := range opts {
		opt(&ret.config)
	}

	if ret.config.DryRun {
		ret.runFunc = func(ctx context.Context, c cloud.Cloud, a Action) ([]Event, error) {
			return a.DryRun(), nil
		}
	} else {
		ret.runFunc = func(ctx context.Context, c cloud.Cloud, a Action) ([]Event, error) {
			return a.Run(ctx, c)
		}
	}

	return ret
}

type serialExecutor struct {
	config  ExecutorConfig
	runFunc func(context.Context, cloud.Cloud, Action) ([]Event, error)
	pending []Action
	done    []Action
	buf     bytes.Buffer
}

func (ex *serialExecutor) String() string { return ex.buf.String() }

func (ex *serialExecutor) Run(ctx context.Context, c cloud.Cloud) ([]Action, error) {
	ex.runEventOnly(ctx, c)

	for a := ex.next(); a != nil; a = ex.next() {
		ex.runAction(ctx, c, a)
	}

	if ex.config.Tracer != nil {
		ex.config.Tracer.Finish(ex.pending)
	}

	return ex.pending, nil
}

func (ex *serialExecutor) runEventOnly(ctx context.Context, c cloud.Cloud) {
	var nonEventOnly []Action
	for _, a := range ex.pending {
		switch a := a.(type) {
		case *eventOnlyAction:
			ex.runAction(ctx, c, a)
		default:
			nonEventOnly = append(nonEventOnly, a)
		}
	}
	ex.pending = nonEventOnly
}

func (ex *serialExecutor) runAction(ctx context.Context, c cloud.Cloud, a Action) {
	te := &TraceEntry{Action: a}
	events, err := ex.runFunc(ctx, c, a)
	if err != nil {
		// TODO
	}
	ex.done = append(ex.done, a)
	for _, ev := range events {
		signaled := ex.signal(ev)
		te.Signaled = append(te.Signaled, signaled...)
	}
	if ex.config.Tracer != nil {
		ex.config.Tracer.Record(te)
	}
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

func (ex *serialExecutor) signal(ev Event) []TraceSignal {
	var ret []TraceSignal
	for _, a := range ex.pending {
		if a.Signal(ev) {
			ret = append(ret, TraceSignal{Event: ev, SignaledAction: a})
		}
	}
	return ret
}
