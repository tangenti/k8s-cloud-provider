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
package lb

import (
	"fmt"

	"github.com/GoogleCloudPlatform/k8s-cloud-provider/pkg/cloud/resgraph"
	"github.com/GoogleCloudPlatform/k8s-cloud-provider/pkg/cloud/resgraph/workflow/testlib"
	"google.golang.org/api/compute/v1"
)

func init() {
	b := resgraph.Builder{Project: "proj"}
	na := resgraph.NodeAttributes{Ownership: resgraph.OwnershipManaged}

	graph1 := func() *resgraph.Graph {
		graph := resgraph.New()

		// Address
		maddr1 := b.N("addr").Address().Resource()
		addr1, _ := maddr1.Freeze()
		graph.AddAddress(addr1, na).SetState(resgraph.NodeExists)

		// ForwardingRule
		mfr1 := b.N("fr").ForwardingRule().Resource()
		mfr1.Access(func(x *compute.ForwardingRule) {
			x.IPAddress = b.N("addr").Address().SelfLink()
			x.Target = b.N("tp").TargetHttpProxy().SelfLink()
		})
		fr1, _ := mfr1.Freeze()
		graph.AddForwardingRule(fr1, na).SetState(resgraph.NodeExists)

		// TargetHttpProxy
		mthp1 := b.N("tp").TargetHttpProxy().Resource()
		mthp1.Access(func(x *compute.TargetHttpProxy) {
			x.UrlMap = b.N("um").UrlMap().SelfLink()
		})
		thp1, _ := mthp1.Freeze()
		graph.AddTargetHttpProxy(thp1, na).SetState(resgraph.NodeExists)

		// UrlMap
		mum1 := b.N("um").UrlMap().Resource()
		mum1.Access(func(x *compute.UrlMap) {
			x.DefaultService = b.N("bs").BackendService().SelfLink()
		})
		um1, _ := mum1.Freeze()
		graph.AddUrlMap(um1, na).SetState(resgraph.NodeExists)

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
		graph.AddBackendService(bs1, na).SetState(resgraph.NodeExists)

		// HealthCheck
		mhc1 := b.N("hc").HealthCheck().Resource()
		hc1, _ := mhc1.Freeze()
		graph.AddHealthCheck(hc1, na).SetState(resgraph.NodeExists)

		// NEG
		mneg1 := b.N("neg").DefaultZone().NetworkEndpointGroup().Resource()
		neg1, _ := mneg1.Freeze()
		graph.AddNetworkEndpointGroup(neg1, na).SetState(resgraph.NodeExists)

		if err := graph.Validate(); err != nil {
			panic(fmt.Sprintf("Graph.Validate = %v, want nil", err))
		}

		return graph
	}

	graph2 := func() *resgraph.Graph {
		graph := resgraph.New()

		// Address
		maddr1 := b.N("addr").Address().Resource()
		maddr1.Access(func(x *compute.Address) {
			x.Description = "changed"
		})
		addr1, _ := maddr1.Freeze()
		graph.AddAddress(addr1, na).SetState(resgraph.NodeExists)

		// ForwardingRule
		mfr1 := b.N("fr").ForwardingRule().Resource()
		mfr1.Access(func(x *compute.ForwardingRule) {
			x.IPAddress = b.N("addr").Address().SelfLink()
			x.Target = b.N("tp").TargetHttpProxy().SelfLink()
		})
		fr1, _ := mfr1.Freeze()
		graph.AddForwardingRule(fr1, na).SetState(resgraph.NodeExists)

		// TargetHttpProxy
		mthp1 := b.N("tp").TargetHttpProxy().Resource()
		mthp1.Access(func(x *compute.TargetHttpProxy) {
			x.UrlMap = b.N("um").UrlMap().SelfLink()
		})
		thp1, _ := mthp1.Freeze()
		graph.AddTargetHttpProxy(thp1, na).SetState(resgraph.NodeExists)

		// UrlMap
		mum1 := b.N("um").UrlMap().Resource()
		mum1.Access(func(x *compute.UrlMap) {
			x.DefaultService = b.N("bs").BackendService().SelfLink()
		})
		um1, _ := mum1.Freeze()
		graph.AddUrlMap(um1, na).SetState(resgraph.NodeExists)

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
		graph.AddBackendService(bs1, na).SetState(resgraph.NodeExists)

		// HealthCheck
		mhc1 := b.N("hc").HealthCheck().Resource()
		hc1, _ := mhc1.Freeze()
		graph.AddHealthCheck(hc1, na).SetState(resgraph.NodeExists)

		// NEG
		mneg1 := b.N("neg").DefaultZone().NetworkEndpointGroup().Resource()
		neg1, _ := mneg1.Freeze()
		graph.AddNetworkEndpointGroup(neg1, na).SetState(resgraph.NodeExists)

		if err := graph.Validate(); err != nil {
			panic(fmt.Sprintf("Graph.Validate = %v, want nil", err))
		}

		return graph
	}

	testlib.Register(&testlib.TestCase{
		Name:        "lb/recreate-addr",
		Description: "Basic LB, recreate Address propagates to the Forwarding Rule.",
		Steps: []testlib.Step{
			{
				Description: "create LB",
				Graph:       graph1(),
			},
			{
				Description: "Address recreation should propagate to Forwarding Rule.",
				Graph:       graph2(),
			},
			{
				Description: "Same graph should have no further updates.",
				Graph:       graph2(),
			},
		},
	})
}
