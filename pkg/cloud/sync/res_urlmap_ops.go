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

type urlMapOps struct{}

func (*urlMapOps) getFuncs(gcp cloud.Cloud) *getFuncs[compute.UrlMap, alpha.UrlMap, beta.UrlMap] {
	return &getFuncs[compute.UrlMap, alpha.UrlMap, beta.UrlMap]{
		ga: getFuncsByScope[compute.UrlMap]{
			global:   gcp.UrlMaps().Get,
			regional: gcp.RegionUrlMaps().Get,
		},
		alpha: getFuncsByScope[alpha.UrlMap]{
			global:   gcp.AlphaUrlMaps().Get,
			regional: gcp.AlphaRegionUrlMaps().Get,
		},
		beta: getFuncsByScope[beta.UrlMap]{
			global:   gcp.BetaUrlMaps().Get,
			regional: gcp.BetaRegionUrlMaps().Get,
		},
	}
}

func (*urlMapOps) createFuncs(gcp cloud.Cloud) *createFuncs[compute.UrlMap, alpha.UrlMap, beta.UrlMap] {
	return &createFuncs[compute.UrlMap, alpha.UrlMap, beta.UrlMap]{
		ga: createFuncsByScope[compute.UrlMap]{
			global:   gcp.UrlMaps().Insert,
			regional: gcp.RegionUrlMaps().Insert,
		},
		alpha: createFuncsByScope[alpha.UrlMap]{
			global:   gcp.AlphaUrlMaps().Insert,
			regional: gcp.AlphaRegionUrlMaps().Insert,
		},
		beta: createFuncsByScope[beta.UrlMap]{
			global:   gcp.BetaUrlMaps().Insert,
			regional: gcp.BetaRegionUrlMaps().Insert,
		},
	}
}

func (*urlMapOps) updateFuncs(gcp cloud.Cloud) *updateFuncs[compute.UrlMap, alpha.UrlMap, beta.UrlMap] {
	return &updateFuncs[compute.UrlMap, alpha.UrlMap, beta.UrlMap]{
		ga: updateFuncsByScope[compute.UrlMap]{
			global:   gcp.UrlMaps().Update,
			regional: gcp.RegionUrlMaps().Update,
		},
		alpha: updateFuncsByScope[alpha.UrlMap]{
			global:   gcp.AlphaUrlMaps().Update,
			regional: gcp.AlphaRegionUrlMaps().Update,
		},
		beta: updateFuncsByScope[beta.UrlMap]{
			global:   gcp.BetaUrlMaps().Update,
			regional: gcp.BetaRegionUrlMaps().Update,
		},
	}
}

func (*urlMapOps) deleteFuncs(gcp cloud.Cloud) *deleteFuncs[compute.UrlMap, alpha.UrlMap, beta.UrlMap] {
	return &deleteFuncs[compute.UrlMap, alpha.UrlMap, beta.UrlMap]{
		ga: deleteFuncsByScope[compute.UrlMap]{
			global:   gcp.UrlMaps().Delete,
			regional: gcp.RegionUrlMaps().Delete,
		},
		alpha: deleteFuncsByScope[alpha.UrlMap]{
			global:   gcp.AlphaUrlMaps().Delete,
			regional: gcp.AlphaRegionUrlMaps().Delete,
		},
		beta: deleteFuncsByScope[beta.UrlMap]{
			global:   gcp.BetaUrlMaps().Delete,
			regional: gcp.BetaRegionUrlMaps().Delete,
		},
	}
}
