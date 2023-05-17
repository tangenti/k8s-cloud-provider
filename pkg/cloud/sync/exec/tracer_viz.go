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
	"fmt"
	"sync/atomic"
)

func NewVizTracer() *VizTracer {
	ret := &VizTracer{}
	return ret
}

type VizTracer struct {
	next atomic.Int32
	buf  bytes.Buffer
}

func (tr *VizTracer) Record(entry *TraceEntry) {
	tr.buf.WriteString(fmt.Sprintf("  \"%s\" [shape=box]\n", entry.Action))
	for _, s := range entry.Signaled {
		tr.buf.WriteString(fmt.Sprintf("  \"%s\" -> \"%s\"\n", entry.Action, s.Event))
		tr.buf.WriteString(fmt.Sprintf("  \"%s\" -> \"%s\"\n", s.Event, s.SignaledAction))
	}
	// TODO: ev needs to be named with trailing edge
}

func (tr *VizTracer) Finish(pending []Action) {
	for _, a := range pending {
		tr.buf.WriteString(fmt.Sprintf("  \"%s\" [style=filled,shape=box,color=pink]\n", a))
		dupe := map[string]struct{}{}
		for _, ev := range a.PendingEvents() {
			if _, ok := dupe[ev.String()]; !ok {
				dupe[ev.String()] = struct{}{}
				tr.buf.WriteString(fmt.Sprintf("  \"%s\" [style=filled,color=pink]\n", ev))
			}
			tr.buf.WriteString(fmt.Sprintf("  \"%s\" -> \"%s\"\n", ev, a))
		}
	}
}

func (tr *VizTracer) String() string {
	var out bytes.Buffer

	out.WriteString("digraph {\n")
	out.WriteString(tr.buf.String())
	out.WriteString("}\n")

	return out.String()
}
