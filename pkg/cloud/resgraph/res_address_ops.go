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

type addressOps struct{}

func (*addressOps) getFuncs(gcp cloud.Cloud) *getFuncs[compute.Address, alpha.Address, beta.Address] {
	return &getFuncs[compute.Address, alpha.Address, beta.Address]{
		ga: getFuncsByScope[compute.Address]{
			global:   gcp.GlobalAddresses().Get,
			regional: gcp.Addresses().Get,
		},
		alpha: getFuncsByScope[alpha.Address]{
			global:   gcp.AlphaGlobalAddresses().Get,
			regional: gcp.AlphaAddresses().Get,
		},
		beta: getFuncsByScope[beta.Address]{
			global:   gcp.BetaGlobalAddresses().Get,
			regional: gcp.BetaAddresses().Get,
		},
	}
}

func (*addressOps) createFuncs(gcp cloud.Cloud) *createFuncs[compute.Address, alpha.Address, beta.Address] {
	return &createFuncs[compute.Address, alpha.Address, beta.Address]{
		ga: createFuncsByScope[compute.Address]{
			global:   gcp.GlobalAddresses().Insert,
			regional: gcp.Addresses().Insert,
		},
		alpha: createFuncsByScope[alpha.Address]{
			global:   gcp.AlphaGlobalAddresses().Insert,
			regional: gcp.AlphaAddresses().Insert,
		},
		beta: createFuncsByScope[beta.Address]{
			global:   gcp.BetaGlobalAddresses().Insert,
			regional: gcp.BetaAddresses().Insert,
		},
	}
}

func (*addressOps) updateFuncs(gcp cloud.Cloud) *updateFuncs[compute.Address, alpha.Address, beta.Address] {
	return nil // Does not support generic Update.
}

func (*addressOps) deleteFuncs(gcp cloud.Cloud) *deleteFuncs[compute.Address, alpha.Address, beta.Address] {
	return &deleteFuncs[compute.Address, alpha.Address, beta.Address]{
		ga: deleteFuncsByScope[compute.Address]{
			global:   gcp.GlobalAddresses().Delete,
			regional: gcp.Addresses().Delete,
		},
		alpha: deleteFuncsByScope[alpha.Address]{
			global:   gcp.AlphaGlobalAddresses().Delete,
			regional: gcp.AlphaAddresses().Delete,
		},
		beta: deleteFuncsByScope[beta.Address]{
			global:   gcp.BetaGlobalAddresses().Delete,
			regional: gcp.BetaAddresses().Delete,
		},
	}
}
