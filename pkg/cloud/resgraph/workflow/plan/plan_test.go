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

package plan

import (
	"context"
	"fmt"
	"testing"

	"github.com/GoogleCloudPlatform/k8s-cloud-provider/pkg/cloud"
	"github.com/GoogleCloudPlatform/k8s-cloud-provider/pkg/cloud/meta"
	"github.com/GoogleCloudPlatform/k8s-cloud-provider/pkg/cloud/resgraph"
	"github.com/GoogleCloudPlatform/k8s-cloud-provider/pkg/cloud/resgraph/algo/graphviz"
	"github.com/GoogleCloudPlatform/k8s-cloud-provider/pkg/cloud/resgraph/exec"
	"google.golang.org/api/compute/v1"
)

func TestLB(t *testing.T) {
	b := resgraph.Builder{Project: "proj"}
	na := resgraph.NodeAttributes{Ownership: resgraph.OwnershipManaged}

	gr := resgraph.New()

	// Address
	maddr1 := b.N("addr").Address().Resource()
	addr1, _ := maddr1.Freeze()
	gr.AddAddress(addr1, na).SetState(resgraph.NodeExists)

	// ForwardingRule
	mfr1 := b.N("fr").ForwardingRule().Resource()
	mfr1.Access(func(x *compute.ForwardingRule) {
		x.IPAddress = b.N("addr").Address().SelfLink()
		x.Target = b.N("tp").TargetHttpProxy().SelfLink()
	})
	fr1, _ := mfr1.Freeze()
	gr.AddForwardingRule(fr1, na).SetState(resgraph.NodeExists)

	// TargetHttpProxy
	mthp1 := b.N("tp").TargetHttpProxy().Resource()
	mthp1.Access(func(x *compute.TargetHttpProxy) {
		x.UrlMap = b.N("um").UrlMap().SelfLink()
	})
	thp1, _ := mthp1.Freeze()
	gr.AddTargetHttpProxy(thp1, na).SetState(resgraph.NodeExists)

	// UrlMap
	mum1 := b.N("um").UrlMap().Resource()
	mum1.Access(func(x *compute.UrlMap) {
		x.DefaultService = b.N("bs").BackendService().SelfLink()
	})
	um1, _ := mum1.Freeze()
	gr.AddUrlMap(um1, na).SetState(resgraph.NodeExists)

	// BackendService
	mbs1 := b.N("bs").BackendService().Resource()
	mbs1.Access(func(x *compute.BackendService) {
		x.Backends = append(x.Backends, &compute.Backend{
			Group: b.N("neg").DefaultZone().NetworkEndpointGroup().SelfLink(),
		})
		x.HealthChecks = []string{
			b.N("hc").HealthCheck().SelfLink(),
		}
	})
	bs1, _ := mbs1.Freeze()
	gr.AddBackendService(bs1, na).SetState(resgraph.NodeExists)

	// HealthCheck
	mhc1 := b.N("hc").HealthCheck().Resource()
	hc1, _ := mhc1.Freeze()
	gr.AddHealthCheck(hc1, na).SetState(resgraph.NodeExists)

	// NEG
	mneg1 := b.N("neg").DefaultZone().NetworkEndpointGroup().Resource()
	neg1, _ := mneg1.Freeze()
	gr.AddNetworkEndpointGroup(neg1, na).SetState(resgraph.NodeExists)

	if err := gr.Validate(); err != nil {
		t.Fatalf("Graph.Validate = %v, want nil", err)
	}

	mock := cloud.NewMockGCE(&cloud.SingleProjectRouter{ID: b.Project})

	mock.HealthChecks().Insert(context.Background(), meta.GlobalKey("hc"), &compute.HealthCheck{})
	mock.BackendServices().Insert(context.Background(), meta.GlobalKey("bs"), &compute.BackendService{Description: "blahblah"})
	mock.TargetHttpProxies().Insert(context.Background(), meta.GlobalKey("tp"), &compute.TargetHttpProxy{
		UrlMap: b.N("umx").UrlMap().SelfLink(),
	})

	mock.UrlMaps().Insert(context.Background(), meta.GlobalKey("umx"), &compute.UrlMap{})
	pl := planner{
		cloud: mock,
		got:   resgraph.New(),
		want:  gr,
	}
	res, err := pl.plan(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	// XXX
	for _, node := range pl.want.All() {
		t.Logf("[PLAN] %v: %s", node.ID(), node.LocalPlan())
	}

	var viz exec.VizTracer
	ex := exec.NewSerialExecutor(res.Actions, exec.DryRunOption(true), exec.TracerOption(&viz))
	pending, err := ex.Run(nil, nil)
	for _, p := range pending {
		t.Logf("%s", p.Summary())
	}
	t.Error(err)
	t.Error(viz.String())

	fmt.Println("got", graphviz.Do(pl.got))
	fmt.Println("want", graphviz.Do(pl.want))
}
