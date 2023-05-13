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

package workflow

import (
	"context"
	"fmt"
	"testing"

	"github.com/GoogleCloudPlatform/k8s-cloud-provider/pkg/cloud"
	"github.com/GoogleCloudPlatform/k8s-cloud-provider/pkg/cloud/meta"
	"github.com/GoogleCloudPlatform/k8s-cloud-provider/pkg/cloud/sync"
	"google.golang.org/api/compute/v1"
)

func TestGC(t *testing.T) {
	const proj = "proj1"
	mock := cloud.NewMockGCE(&cloud.SingleProjectRouter{ID: proj})

	// Setup the most basic LB resource tree.
	mock.GlobalForwardingRules().Insert(context.Background(), meta.GlobalKey("fr1"), &compute.ForwardingRule{
		Name:   "fr1",
		Target: cloud.SelfLink(meta.VersionGA, proj, "targetHttpProxies", meta.GlobalKey("tp1")),
	})
	mock.TargetHttpProxies().Insert(context.Background(), meta.GlobalKey("tp1"), &compute.TargetHttpProxy{
		Name:   "tp1",
		UrlMap: cloud.SelfLink(meta.VersionGA, proj, "urlMaps", meta.GlobalKey("um1")),
	})
	mock.UrlMaps().Insert(context.Background(), meta.GlobalKey("um1"), &compute.UrlMap{
		Name:           "um1",
		DefaultService: cloud.SelfLink(meta.VersionGA, proj, "backendServices", meta.GlobalKey("bs1")),
	})
	mock.BackendServices().Insert(context.Background(), meta.GlobalKey("bs1"), &compute.BackendService{
		Name: "bs1",
		Backends: []*compute.Backend{
			{
				Group: cloud.SelfLink(meta.VersionGA, proj, "networkEndpointGroups", meta.ZonalKey("neg1", "us-central1-b")),
			},
		},
	})
	mock.NetworkEndpointGroups().Insert(context.Background(), meta.ZonalKey("neg1", "us-central1-b"), &compute.NetworkEndpointGroup{
		Name: "neg1",
	})

	err := GarbageCollect(context.Background(), mock, &GarbageCollectConfig{
		Scope:          GarbageCollectGlobal,
		ResourcePrefix: "",
		NodeIsLive: func(node sync.Node) bool {
			switch node := node.(type) {
			case *sync.UrlMapNode:
				// TODO
				fmt.Println(node)
			}
			return false
		},
	})
	t.Error(err)
}
