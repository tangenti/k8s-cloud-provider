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

type healthCheckOps struct{}

func (*healthCheckOps) getFuncs(gcp cloud.Cloud) *getFuncs[compute.HealthCheck, alpha.HealthCheck, beta.HealthCheck] {
	return &getFuncs[compute.HealthCheck, alpha.HealthCheck, beta.HealthCheck]{
		ga: getFuncsByScope[compute.HealthCheck]{
			global:   gcp.HealthChecks().Get,
			regional: gcp.RegionHealthChecks().Get,
		},
		alpha: getFuncsByScope[alpha.HealthCheck]{
			global:   gcp.AlphaHealthChecks().Get,
			regional: gcp.AlphaRegionHealthChecks().Get,
		},
		beta: getFuncsByScope[beta.HealthCheck]{
			global:   gcp.BetaHealthChecks().Get,
			regional: gcp.BetaRegionHealthChecks().Get,
		},
	}
}

func (*healthCheckOps) createFuncs(gcp cloud.Cloud) *createFuncs[compute.HealthCheck, alpha.HealthCheck, beta.HealthCheck] {
	return &createFuncs[compute.HealthCheck, alpha.HealthCheck, beta.HealthCheck]{
		ga: createFuncsByScope[compute.HealthCheck]{
			global:   gcp.HealthChecks().Insert,
			regional: gcp.RegionHealthChecks().Insert,
		},
		alpha: createFuncsByScope[alpha.HealthCheck]{
			global:   gcp.AlphaHealthChecks().Insert,
			regional: gcp.AlphaRegionHealthChecks().Insert,
		},
		beta: createFuncsByScope[beta.HealthCheck]{
			global:   gcp.BetaHealthChecks().Insert,
			regional: gcp.BetaRegionHealthChecks().Insert,
		},
	}
}

func (*healthCheckOps) updateFuncs(gcp cloud.Cloud) *updateFuncs[compute.HealthCheck, alpha.HealthCheck, beta.HealthCheck] {
	return &updateFuncs[compute.HealthCheck, alpha.HealthCheck, beta.HealthCheck]{
		ga: updateFuncsByScope[compute.HealthCheck]{
			global:   gcp.HealthChecks().Insert,
			regional: gcp.RegionHealthChecks().Update,
		},
		alpha: updateFuncsByScope[alpha.HealthCheck]{
			global:   gcp.AlphaHealthChecks().Insert,
			regional: gcp.AlphaRegionHealthChecks().Update,
		},
		beta: updateFuncsByScope[beta.HealthCheck]{
			global:   gcp.BetaHealthChecks().Insert,
			regional: gcp.BetaRegionHealthChecks().Update,
		},
		options: updateFuncsNoFingerprint,
	}
}

func (*healthCheckOps) deleteFuncs(gcp cloud.Cloud) *deleteFuncs[compute.HealthCheck, alpha.HealthCheck, beta.HealthCheck] {
	return &deleteFuncs[compute.HealthCheck, alpha.HealthCheck, beta.HealthCheck]{
		ga: deleteFuncsByScope[compute.HealthCheck]{
			global:   gcp.HealthChecks().Delete,
			regional: gcp.RegionHealthChecks().Delete,
		},
		alpha: deleteFuncsByScope[alpha.HealthCheck]{
			global:   gcp.AlphaHealthChecks().Delete,
			regional: gcp.AlphaRegionHealthChecks().Delete,
		},
		beta: deleteFuncsByScope[beta.HealthCheck]{
			global:   gcp.BetaHealthChecks().Delete,
			regional: gcp.BetaRegionHealthChecks().Delete,
		},
	}
}
