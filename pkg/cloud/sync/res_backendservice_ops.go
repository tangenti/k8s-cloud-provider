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

type backendServiceOps struct{}

func (*backendServiceOps) getFuncs(gcp cloud.Cloud) *getFuncs[compute.BackendService, alpha.BackendService, beta.BackendService] {
	return &getFuncs[compute.BackendService, alpha.BackendService, beta.BackendService]{
		ga: getFuncsByScope[compute.BackendService]{
			global:   gcp.BackendServices().Get,
			regional: gcp.RegionBackendServices().Get,
		},
		alpha: getFuncsByScope[alpha.BackendService]{
			global:   gcp.AlphaBackendServices().Get,
			regional: gcp.AlphaRegionBackendServices().Get,
		},
		beta: getFuncsByScope[beta.BackendService]{
			global:   gcp.BetaBackendServices().Get,
			regional: gcp.BetaRegionBackendServices().Get,
		},
	}
}

func (*backendServiceOps) createFuncs(gcp cloud.Cloud) *createFuncs[compute.BackendService, alpha.BackendService, beta.BackendService] {
	return &createFuncs[compute.BackendService, alpha.BackendService, beta.BackendService]{
		ga: createFuncsByScope[compute.BackendService]{
			global:   gcp.BackendServices().Insert,
			regional: gcp.RegionBackendServices().Insert,
		},
		alpha: createFuncsByScope[alpha.BackendService]{
			global:   gcp.AlphaBackendServices().Insert,
			regional: gcp.AlphaRegionBackendServices().Insert,
		},
		beta: createFuncsByScope[beta.BackendService]{
			global:   gcp.BetaBackendServices().Insert,
			regional: gcp.BetaRegionBackendServices().Insert,
		},
	}
}

func (*backendServiceOps) updateFuncs(gcp cloud.Cloud) *updateFuncs[compute.BackendService, alpha.BackendService, beta.BackendService] {
	return &updateFuncs[compute.BackendService, alpha.BackendService, beta.BackendService]{
		ga: updateFuncsByScope[compute.BackendService]{
			global:   gcp.BackendServices().Update,
			regional: gcp.RegionBackendServices().Update,
		},
		alpha: updateFuncsByScope[alpha.BackendService]{
			global:   gcp.AlphaBackendServices().Update,
			regional: gcp.AlphaRegionBackendServices().Update,
		},
		beta: updateFuncsByScope[beta.BackendService]{
			global:   gcp.BetaBackendServices().Update,
			regional: gcp.BetaRegionBackendServices().Update,
		},
	}
}

func (*backendServiceOps) deleteFuncs(gcp cloud.Cloud) *deleteFuncs[compute.BackendService, alpha.BackendService, beta.BackendService] {
	return &deleteFuncs[compute.BackendService, alpha.BackendService, beta.BackendService]{
		ga: deleteFuncsByScope[compute.BackendService]{
			global:   gcp.BackendServices().Delete,
			regional: gcp.RegionBackendServices().Delete,
		},
		alpha: deleteFuncsByScope[alpha.BackendService]{
			global:   gcp.AlphaBackendServices().Delete,
			regional: gcp.AlphaRegionBackendServices().Delete,
		},
		beta: deleteFuncsByScope[beta.BackendService]{
			global:   gcp.BetaBackendServices().Delete,
			regional: gcp.BetaRegionBackendServices().Delete,
		},
	}
}
