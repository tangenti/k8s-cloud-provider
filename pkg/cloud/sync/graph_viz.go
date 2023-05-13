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

	"github.com/GoogleCloudPlatform/k8s-cloud-provider/pkg/cloud"
	"github.com/GoogleCloudPlatform/k8s-cloud-provider/pkg/cloud/meta"
)

// Graphviz dumps out a graphviz representation of the resource graph.
func Graphviz(g *Graph) string {
	pp := func(id *cloud.ResourceID) string {
		var k string
		switch id.Key.Type() {
		case meta.Global:
			k = id.Key.Name
		case meta.Regional:
			k = id.Key.Region + "_" + id.Key.Name
		case meta.Zonal:
			k = id.Key.Zone + "_" + id.Key.Name
		}
		return fmt.Sprintf(`"%s\n%s\n%s"`, id.Resource, id.ProjectID, k)
	}

	var buf bytes.Buffer

	buf.WriteString("digraph G {\n")

	for mk, v := range g.all {
		k := mk.ToID()
		buf.WriteString(pp(k) + "\n")
		deps, _ := v.OutRefs() // TODO:err
		for _, dep := range deps {
			buf.WriteString(fmt.Sprintf("  %s -> %s\n", pp(k), pp(dep.To)))
		}
	}
	buf.WriteString("}\n")

	return buf.String()
}
