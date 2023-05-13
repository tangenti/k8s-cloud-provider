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
	"github.com/GoogleCloudPlatform/k8s-cloud-provider/pkg/cloud"
	alpha "google.golang.org/api/compute/v0.alpha"
	beta "google.golang.org/api/compute/v0.beta"
	"google.golang.org/api/compute/v1"
)

type targetHttpProxyOps struct{}

func (*targetHttpProxyOps) getFuncs(gcp cloud.Cloud) *getFuncs[compute.TargetHttpProxy, alpha.TargetHttpProxy, beta.TargetHttpProxy] {
	return &getFuncs[compute.TargetHttpProxy, alpha.TargetHttpProxy, beta.TargetHttpProxy]{
		ga: getFuncsByScope[compute.TargetHttpProxy]{
			global:   gcp.TargetHttpProxies().Get,
			regional: gcp.RegionTargetHttpProxies().Get,
		},
		alpha: getFuncsByScope[alpha.TargetHttpProxy]{
			global:   gcp.AlphaTargetHttpProxies().Get,
			regional: gcp.AlphaRegionTargetHttpProxies().Get,
		},
		beta: getFuncsByScope[beta.TargetHttpProxy]{
			global:   gcp.BetaTargetHttpProxies().Get,
			regional: gcp.BetaRegionTargetHttpProxies().Get,
		},
	}
}

func (*targetHttpProxyOps) createFuncs(gcp cloud.Cloud) *createFuncs[compute.TargetHttpProxy, alpha.TargetHttpProxy, beta.TargetHttpProxy] {
	return &createFuncs[compute.TargetHttpProxy, alpha.TargetHttpProxy, beta.TargetHttpProxy]{
		ga: createFuncsByScope[compute.TargetHttpProxy]{
			global:   gcp.TargetHttpProxies().Insert,
			regional: gcp.RegionTargetHttpProxies().Insert,
		},
		alpha: createFuncsByScope[alpha.TargetHttpProxy]{
			global:   gcp.AlphaTargetHttpProxies().Insert,
			regional: gcp.AlphaRegionTargetHttpProxies().Insert,
		},
		beta: createFuncsByScope[beta.TargetHttpProxy]{
			global:   gcp.BetaTargetHttpProxies().Insert,
			regional: gcp.BetaRegionTargetHttpProxies().Insert,
		},
	}
}

func (*targetHttpProxyOps) updateFuncs(gcp cloud.Cloud) *updateFuncs[compute.TargetHttpProxy, alpha.TargetHttpProxy, beta.TargetHttpProxy] {
	return nil // Does not support generic Update.
}

func (*targetHttpProxyOps) deleteFuncs(gcp cloud.Cloud) *deleteFuncs[compute.TargetHttpProxy, alpha.TargetHttpProxy, beta.TargetHttpProxy] {
	return &deleteFuncs[compute.TargetHttpProxy, alpha.TargetHttpProxy, beta.TargetHttpProxy]{
		ga: deleteFuncsByScope[compute.TargetHttpProxy]{
			global:   gcp.TargetHttpProxies().Delete,
			regional: gcp.RegionTargetHttpProxies().Delete,
		},
		alpha: deleteFuncsByScope[alpha.TargetHttpProxy]{
			global:   gcp.AlphaTargetHttpProxies().Delete,
			regional: gcp.AlphaRegionTargetHttpProxies().Delete,
		},
		beta: deleteFuncsByScope[beta.TargetHttpProxy]{
			global:   gcp.BetaTargetHttpProxies().Delete,
			regional: gcp.BetaRegionTargetHttpProxies().Delete,
		},
	}
}
