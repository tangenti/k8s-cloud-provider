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
	"bytes"
	"fmt"
	"sort"
)

// Graphviz returns a .dot (http://graphviz.org) representation of the resource
// graph for visualization.
func Graphviz(g *Graph) string {
	var buf bytes.Buffer
	buf.WriteString("digraph G {\n")
	buf.WriteString("  rankdir=TB\n") // layout top to bottom.

	for mk, v := range g.all {
		k := mk.ToID()
		gn := &graphvizNode{
			name:  k.String(),
			shape: "box",
			style: "filled",
			kv: map[string]any{
				"localPlan": v.LocalPlan().GraphvizString(),
				"state":     v.State(),
				"version":   v.version(),
			},
		}
		deps, outRefErr := v.OutRefs()
		for _, dep := range deps {
			buf.WriteString(fmt.Sprintf(`  "%s" -> "%s"`+"\n", k, dep.To))
		}

		gn.color = gn.opColor(v.LocalPlan().Op())

		// errors
		if v.GetErr() != nil || outRefErr != nil {
			var errStr string
			if v.GetErr() != nil {
				errStr += fmt.Sprintf("GetErr()=%v ", v.GetErr())
			}
			if outRefErr != nil {
				errStr += fmt.Sprintf("OutRefs()=%v ", outRefErr)
			}
			gn.kv["errors"] = errStr
		}
		buf.WriteString(gn.String())
	}
	buf.WriteString("}\n")

	return buf.String()
}

type graphvizNode struct {
	name string

	color string
	shape string
	style string

	kv map[string]any
}

func (*graphvizNode) indent(n int) string {
	var ret string
	for i := 0; i < n; i++ {
		ret += "  "
	}
	return ret
}

func (*graphvizNode) opColor(op Operation) string {
	switch op {
	case OpCreate:
		return "chartreuse"
	case OpDelete:
		return "lightcoral"
	case OpRecreate:
		return "orange"
	case OpUpdate:
		return "khaki1"
	case OpNothing:
		return "beige"
	case OpUnknown:
		return "red"
	}
	return "mediumpurple1"
}

func (n *graphvizNode) String() string {
	type line struct {
		indent int
		s      string
	}

	var lines []line

	lines = append(lines, line{1, fmt.Sprintf("\"%s\" [label=<", n.name)})
	lines = append(lines, line{2, "<table border=\"0\">"})
	lines = append(lines, line{3, "<tr><td colspan=\"2\"><font point-size=\"16\">\\N</font></td></tr>"})
	lines = append(lines, line{3, "<tr><td colspan=\"2\">---</td></tr>"})

	var keys []string
	for k := range n.kv {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		lines = append(lines, line{3, fmt.Sprintf("<tr><td>%s</td><td align=\"left\">%v</td></tr>", k, n.kv[k])})
	}
	lines = append(lines, line{2, "</table>"})

	var attribsStr string
	for _, at := range []struct {
		key string
		val *string
	}{
		{"color", &n.color},
		{"shape", &n.shape},
		{"style", &n.style},
	} {
		if *at.val != "" {
			attribsStr += fmt.Sprintf(`,%s=%s`, at.key, *at.val)
		}
	}
	lines = append(lines, line{1, fmt.Sprintf(">%s]", attribsStr)})

	var out string
	for _, l := range lines {
		out += n.indent(l.indent) + l.s + "\n"
	}
	return out
}
