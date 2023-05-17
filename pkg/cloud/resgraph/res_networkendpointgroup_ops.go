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
	"github.com/GoogleCloudPlatform/k8s-cloud-provider/pkg/cloud"
	alpha "google.golang.org/api/compute/v0.alpha"
	beta "google.golang.org/api/compute/v0.beta"
	"google.golang.org/api/compute/v1"
)

type networkEndpointGroupOps struct{}

func (*networkEndpointGroupOps) getFuncs(gcp cloud.Cloud) *getFuncs[compute.NetworkEndpointGroup, alpha.NetworkEndpointGroup, beta.NetworkEndpointGroup] {
	return &getFuncs[compute.NetworkEndpointGroup, alpha.NetworkEndpointGroup, beta.NetworkEndpointGroup]{
		ga: getFuncsByScope[compute.NetworkEndpointGroup]{
			zonal: gcp.NetworkEndpointGroups().Get,
		},
		alpha: getFuncsByScope[alpha.NetworkEndpointGroup]{
			zonal: gcp.AlphaNetworkEndpointGroups().Get,
		},
		beta: getFuncsByScope[beta.NetworkEndpointGroup]{
			zonal: gcp.BetaNetworkEndpointGroups().Get,
		},
	}
}

func (*networkEndpointGroupOps) createFuncs(gcp cloud.Cloud) *createFuncs[compute.NetworkEndpointGroup, alpha.NetworkEndpointGroup, beta.NetworkEndpointGroup] {
	return &createFuncs[compute.NetworkEndpointGroup, alpha.NetworkEndpointGroup, beta.NetworkEndpointGroup]{
		ga: createFuncsByScope[compute.NetworkEndpointGroup]{
			zonal: gcp.NetworkEndpointGroups().Insert,
		},
		alpha: createFuncsByScope[alpha.NetworkEndpointGroup]{
			zonal: gcp.AlphaNetworkEndpointGroups().Insert,
		},
		beta: createFuncsByScope[beta.NetworkEndpointGroup]{
			zonal: gcp.BetaNetworkEndpointGroups().Insert,
		},
	}
}

func (*networkEndpointGroupOps) updateFuncs(gcp cloud.Cloud) *updateFuncs[compute.NetworkEndpointGroup, alpha.NetworkEndpointGroup, beta.NetworkEndpointGroup] {
	return nil // Does not support generic Update.
}

func (*networkEndpointGroupOps) deleteFuncs(gcp cloud.Cloud) *deleteFuncs[compute.NetworkEndpointGroup, alpha.NetworkEndpointGroup, beta.NetworkEndpointGroup] {
	return &deleteFuncs[compute.NetworkEndpointGroup, alpha.NetworkEndpointGroup, beta.NetworkEndpointGroup]{
		ga: deleteFuncsByScope[compute.NetworkEndpointGroup]{
			zonal: gcp.NetworkEndpointGroups().Delete,
		},
		alpha: deleteFuncsByScope[alpha.NetworkEndpointGroup]{
			zonal: gcp.AlphaNetworkEndpointGroups().Delete,
		},
		beta: deleteFuncsByScope[beta.NetworkEndpointGroup]{
			zonal: gcp.BetaNetworkEndpointGroups().Delete,
		},
	}
}
