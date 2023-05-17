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

import "testing"

func TestExecutor(t *testing.T) {
	var viz VizTracer
	exec := NewSerialExecutor([]Action{
		newTestAction("a", []Event{}, []Event{StringEvent("be"), StringEvent("ce")}),
		newTestAction("b", []Event{StringEvent("be")}, []Event{StringEvent("de1")}),
		newTestAction("c", []Event{StringEvent("ce")}, []Event{StringEvent("de2")}),
		newTestAction("d", []Event{StringEvent("de1"), StringEvent("de2")}, []Event{}),
	}, DryRunOption(true), TracerOption(&viz))
	t.Error(exec.Run(nil, nil))
	t.Error(viz.String())
}
