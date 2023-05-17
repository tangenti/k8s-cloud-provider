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
	"sort"
	"testing"

	"github.com/GoogleCloudPlatform/k8s-cloud-provider/pkg/cloud"
	"github.com/GoogleCloudPlatform/k8s-cloud-provider/pkg/cloud/meta"
	"github.com/google/go-cmp/cmp"
)

func TestGraphInRefs(t *testing.T) {
	const project = "proj1"

	var (
		aID = fakeID(project, meta.GlobalKey("a"))
		bID = fakeID(project, meta.GlobalKey("b"))
		cID = fakeID(project, meta.GlobalKey("c"))
		dID = fakeID(project, meta.GlobalKey("d"))
	)

	g := New()

	sl := func(id *cloud.ResourceID) string { return id.SelfLink(meta.VersionGA) }
	addFake := func(name string, deps []string) {
		id := fakeID(project, meta.GlobalKey(name))
		mf := newMutableFake(project, id.Key)
		mf.Access(func(x *fakeResource) {
			for _, d := range deps {
				dID := &cloud.ResourceID{Resource: "fakes", ProjectID: project, Key: meta.GlobalKey(d)}
				x.Dependencies = append(x.Dependencies, sl(dID))
			}
		})
		f, _ := mf.Freeze()
		g.addFake(f, NodeAttributes{Ownership: OwnershipManaged})
	}

	addFake("a", []string{"b", "c"})
	addFake("b", []string{"d"})
	addFake("c", []string{"b", "d"})
	addFake("d", []string{})

	if err := g.Validate(); err != nil {
		t.Fatalf("Validate() = %v, want nil", err)
	}

	keys := func(nl []Node) []string {
		var ret []string
		for _, n := range nl {
			ret = append(ret, sl(n.ID()))
		}
		sort.Strings(ret)
		return ret
	}

	for _, want := range []struct {
		key  *cloud.ResourceID
		deps []string
	}{
		{aID, nil},
		{bID, []string{sl(aID), sl(cID)}},
		{cID, []string{sl(aID)}},
		{dID, []string{sl(bID), sl(cID)}},
	} {
		g.ComputeInRefs()

		node := g.Resource(want.key)
		if node == nil {
			t.Fatalf("Node %s should exist", want.key)
		}
		var nodes []Node
		for _, r := range node.InRefs() {
			nodes = append(nodes, g.Resource(r.From))
		}
		t.Logf("nodes = %+v", nodes)
		got := keys(nodes)
		if diff := cmp.Diff(got, want.deps); diff != "" {
			t.Errorf("Dependents(%s), -got,+want: %s", want.key, diff)
		}
	}
}
