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
	"fmt"

	"github.com/GoogleCloudPlatform/k8s-cloud-provider/pkg/cloud"
)

func NewNodeByID(id *cloud.ResourceID, na NodeAttributes) (Node, error) {
	var node Node

	switch id.Resource {
	case "addresses":
		node = newAddressNode(id)
	case "backendServices":
		node = newBackendServiceNode(id)
	case "forwardingRules":
		node = newForwardingRuleNode(id)
	case "healthChecks":
		node = newHealthCheckNode(id)
	case "networkEndpointGroups":
		node = newNetworkEndpointGroupNode(id)
	case "targetHttpProxies":
		node = newTargetHttpProxyNode(id)
	case "urlMaps":
		node = newUrlMapNode(id)
	default:
		return nil, fmt.Errorf("NewNodeByResourceID: unknown resource type %q", id.Resource)
	}

	node.setAttributes(na)

	return node, nil
}

func ResourceOutRefs(res any) ([]ResourceRef, error) {
	switch res := res.(type) {
	case Address:
		return AddressOutRefs(res)
	case BackendService:
		return BackendServiceOutRefs(res)
	case ForwardingRule:
		return ForwardingRuleOutRefs(res)
	case HealthCheck:
		return HealthCheckOutRefs(res)
	case NetworkEndpointGroup:
		return NetworkEndpointGroupOutRefs(res)
	case TargetHttpProxy:
		return TargetHttpProxyOutRefs(res)
	case UrlMap:
		return UrlMapOutRefs(res)
	}
	return nil, fmt.Errorf("ResourceOutRefs: invalid type %T", res)
}
