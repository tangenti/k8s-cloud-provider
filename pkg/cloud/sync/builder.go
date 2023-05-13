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
	"github.com/GoogleCloudPlatform/k8s-cloud-provider/pkg/cloud/meta"
)

// Builder is a convenience wrapper for tests. Do not use this in production.
type Builder struct {
	Project string
	Name    string
	Region  string
	Zone    string
}

func (b *Builder) Key() *meta.Key {
	switch {
	case b.Region == "" && b.Zone == "":
		return meta.GlobalKey(b.Name)
	case b.Region != "":
		return meta.RegionalKey(b.Name, b.Region)
	case b.Zone != "":
		return meta.ZonalKey(b.Name, b.Zone)
	}
	panic(fmt.Sprintf("missing fields: %+v", *b))
}

func (b *Builder) P(project string) *Builder {
	ret := *b
	ret.Project = project
	return &ret
}

func (b *Builder) N(name string) *Builder {
	ret := *b
	ret.Name = name
	return &ret
}

func (b *Builder) R(region string) *Builder {
	ret := *b
	ret.Region = region
	return &ret
}

func (b *Builder) Z(zone string) *Builder {
	ret := *b
	ret.Zone = zone
	return &ret
}

func (b *Builder) DefaultRegion() *Builder { return b.R("us-central1") }
func (b *Builder) DefaultZone() *Builder   { return b.Z("us-central1-b") }

func (b *Builder) Address() *AddressBuilder               { return &AddressBuilder{*b} }
func (b *Builder) BackendService() *BackendServiceBuilder { return &BackendServiceBuilder{*b} }
func (b *Builder) ForwardingRule() *ForwardingRuleBuilder { return &ForwardingRuleBuilder{*b} }
func (b *Builder) HealthCheck() *HealthCheckBuilder       { return &HealthCheckBuilder{*b} }
func (b *Builder) NetworkEndpointGroup() *NetworkEndpointGroupBuilder {
	return &NetworkEndpointGroupBuilder{*b}
}
func (b *Builder) TargetHttpProxy() *TargetHttpProxyBuilder { return &TargetHttpProxyBuilder{*b} }
func (b *Builder) UrlMap() *UrlMapBuilder                   { return &UrlMapBuilder{*b} }

type AddressBuilder struct{ Builder }

func (b *AddressBuilder) ID() *cloud.ResourceID { return AddressID(b.Project, b.Key()) }
func (b *AddressBuilder) SelfLink() string      { return b.ID().SelfLink(meta.VersionGA) }
func (b *AddressBuilder) Resource() MutableAddress {
	return NewMutableAddress(b.Project, b.Key())
}
func (b *AddressBuilder) Node() *AddressNode {
	return newAddressNode(b.ID())
}

type BackendServiceBuilder struct{ Builder }

func (b *BackendServiceBuilder) ID() *cloud.ResourceID {
	return BackendServiceID(b.Project, b.Key())
}
func (b *BackendServiceBuilder) SelfLink() string { return b.ID().SelfLink(meta.VersionGA) }
func (b *BackendServiceBuilder) Resource() MutableBackendService {
	return NewMutableBackendService(b.Project, b.Key())
}
func (b *BackendServiceBuilder) Node() *BackendServiceNode {
	return newBackendServiceNode(b.ID())
}

type ForwardingRuleBuilder struct{ Builder }

func (b *ForwardingRuleBuilder) ID() *cloud.ResourceID {
	return ForwardingRuleID(b.Project, b.Key())
}
func (b *ForwardingRuleBuilder) SelfLink() string { return b.ID().SelfLink(meta.VersionGA) }
func (b *ForwardingRuleBuilder) Resource() MutableForwardingRule {
	return NewMutableForwardingRule(b.Project, b.Key())
}
func (b *ForwardingRuleBuilder) Node() *ForwardingRuleNode {
	return newForwardingRuleNode(b.ID())
}

type HealthCheckBuilder struct{ Builder }

func (b *HealthCheckBuilder) ID() *cloud.ResourceID { return HealthCheckID(b.Project, b.Key()) }
func (b *HealthCheckBuilder) SelfLink() string      { return b.ID().SelfLink(meta.VersionGA) }
func (b *HealthCheckBuilder) Resource() MutableHealthCheck {
	return NewMutableHealthCheck(b.Project, b.Key())
}
func (b *HealthCheckBuilder) Node() *HealthCheckNode {
	return newHealthCheckNode(b.ID())
}

type NetworkEndpointGroupBuilder struct{ Builder }

func (b *NetworkEndpointGroupBuilder) ID() *cloud.ResourceID {
	return NetworkEndpointGroupID(b.Project, b.Key())
}
func (b *NetworkEndpointGroupBuilder) SelfLink() string { return b.ID().SelfLink(meta.VersionGA) }
func (b *NetworkEndpointGroupBuilder) Resource() MutableNetworkEndpointGroup {
	return NewMutableNetworkEndpointGroup(b.Project, b.Key())
}
func (b *NetworkEndpointGroupBuilder) Node() *NetworkEndpointGroupNode {
	return newNetworkEndpointGroupNode(b.ID())
}

type TargetHttpProxyBuilder struct{ Builder }

func (b *TargetHttpProxyBuilder) ID() *cloud.ResourceID {
	return TargetHttpProxyID(b.Project, b.Key())
}
func (b *TargetHttpProxyBuilder) SelfLink() string { return b.ID().SelfLink(meta.VersionGA) }
func (b *TargetHttpProxyBuilder) Resource() MutableTargetHttpProxy {
	return NewMutableTargetHttpProxy(b.Project, b.Key())
}
func (b *TargetHttpProxyBuilder) Node() *TargetHttpProxyNode {
	return newTargetHttpProxyNode(b.ID())
}

type UrlMapBuilder struct{ Builder }

func (b *UrlMapBuilder) ID() *cloud.ResourceID { return UrlMapID(b.Project, b.Key()) }
func (b *UrlMapBuilder) SelfLink() string      { return b.ID().SelfLink(meta.VersionGA) }
func (b *UrlMapBuilder) Resource() MutableUrlMap {
	return NewMutableUrlMap(b.Project, b.Key())
}
func (b *UrlMapBuilder) Node() *UrlMapNode {
	return newUrlMapNode(b.ID())
}
