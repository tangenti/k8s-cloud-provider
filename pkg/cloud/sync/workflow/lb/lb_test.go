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
	"context"
	"testing"

	"github.com/GoogleCloudPlatform/k8s-cloud-provider/pkg/cloud"
	"github.com/GoogleCloudPlatform/k8s-cloud-provider/pkg/cloud/meta"
	"github.com/GoogleCloudPlatform/k8s-cloud-provider/pkg/cloud/sync"
	"google.golang.org/api/compute/v1"
)

func TestLB(t *testing.T) {
	b := sync.Builder{Project: "proj"}
	na := sync.NodeAttributes{Ownership: sync.OwnershipManaged}

	graph := sync.NewGraph()

	// Address
	maddr1 := b.N("addr").Address().Resource()
	addr1, _ := maddr1.Freeze()
	graph.AddAddress(addr1, na).SetState(sync.NodeStateExists)

	// ForwardingRule
	mfr1 := b.N("fr").ForwardingRule().Resource()
	mfr1.Access(func(x *compute.ForwardingRule) {
		x.IPAddress = b.N("addr").Address().SelfLink()
		x.Target = b.N("tp").TargetHttpProxy().SelfLink()
	})
	fr1, _ := mfr1.Freeze()
	graph.AddForwardingRule(fr1, na).SetState(sync.NodeStateExists)

	// TargetHttpProxy
	mthp1 := b.N("tp").TargetHttpProxy().Resource()
	mthp1.Access(func(x *compute.TargetHttpProxy) {
		x.UrlMap = b.N("um").UrlMap().SelfLink()
	})
	thp1, _ := mthp1.Freeze()
	graph.AddTargetHttpProxy(thp1, na).SetState(sync.NodeStateExists)

	// UrlMap
	mum1 := b.N("um").UrlMap().Resource()
	mum1.Access(func(x *compute.UrlMap) {
		x.DefaultService = b.N("bs").BackendService().SelfLink()
	})
	um1, _ := mum1.Freeze()
	graph.AddUrlMap(um1, na).SetState(sync.NodeStateExists)

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
	graph.AddBackendService(bs1, na).SetState(sync.NodeStateExists)

	// HealthCheck
	mhc1 := b.N("hc").HealthCheck().Resource()
	hc1, _ := mhc1.Freeze()
	graph.AddHealthCheck(hc1, na).SetState(sync.NodeStateExists)

	// NEG
	mneg1 := b.N("neg").DefaultZone().NetworkEndpointGroup().Resource()
	neg1, _ := mneg1.Freeze()
	graph.AddNetworkEndpointGroup(neg1, na).SetState(sync.NodeStateExists)

	if err := graph.Validate(); err != nil {
		t.Fatalf("Graph.Validate = %v, want nil", err)
	}

	mock := cloud.NewMockGCE(&cloud.SingleProjectRouter{ID: b.Project})

	mock.HealthChecks().Insert(context.Background(), meta.GlobalKey("hc"), &compute.HealthCheck{})
	mock.BackendServices().Insert(context.Background(), meta.GlobalKey("bs"), &compute.BackendService{})
	mock.TargetHttpProxies().Insert(context.Background(), meta.GlobalKey("tp"), &compute.TargetHttpProxy{
		UrlMap: b.N("um").UrlMap().SelfLink(),
	})

	err := Workflow(context.Background(), mock, graph)
	if err != nil {
		t.Fatalf("Workflow = %v, want nil", err)
	}

	t.Error("x")
}
