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

type forwardingRuleOps struct{}

func (*forwardingRuleOps) getFuncs(gcp cloud.Cloud) *getFuncs[compute.ForwardingRule, alpha.ForwardingRule, beta.ForwardingRule] {
	return &getFuncs[compute.ForwardingRule, alpha.ForwardingRule, beta.ForwardingRule]{
		ga: getFuncsByScope[compute.ForwardingRule]{
			global:   gcp.GlobalForwardingRules().Get,
			regional: gcp.ForwardingRules().Get,
		},
		alpha: getFuncsByScope[alpha.ForwardingRule]{
			global:   gcp.AlphaGlobalForwardingRules().Get,
			regional: gcp.AlphaForwardingRules().Get,
		},
		beta: getFuncsByScope[beta.ForwardingRule]{
			global:   gcp.BetaGlobalForwardingRules().Get,
			regional: gcp.BetaForwardingRules().Get,
		},
	}
}

func (*forwardingRuleOps) createFuncs(gcp cloud.Cloud) *createFuncs[compute.ForwardingRule, alpha.ForwardingRule, beta.ForwardingRule] {
	return &createFuncs[compute.ForwardingRule, alpha.ForwardingRule, beta.ForwardingRule]{
		ga: createFuncsByScope[compute.ForwardingRule]{
			global:   gcp.GlobalForwardingRules().Insert,
			regional: gcp.ForwardingRules().Insert,
		},
		alpha: createFuncsByScope[alpha.ForwardingRule]{
			global:   gcp.AlphaGlobalForwardingRules().Insert,
			regional: gcp.AlphaForwardingRules().Insert,
		},
		beta: createFuncsByScope[beta.ForwardingRule]{
			global:   gcp.BetaGlobalForwardingRules().Insert,
			regional: gcp.BetaForwardingRules().Insert,
		},
	}
}

func (*forwardingRuleOps) updateFuncs(cloud.Cloud) *updateFuncs[compute.ForwardingRule, alpha.ForwardingRule, beta.ForwardingRule] {
	return nil // Does not support generic Update.
}

func (*forwardingRuleOps) deleteFuncs(gcp cloud.Cloud) *deleteFuncs[compute.ForwardingRule, alpha.ForwardingRule, beta.ForwardingRule] {
	return &deleteFuncs[compute.ForwardingRule, alpha.ForwardingRule, beta.ForwardingRule]{
		ga: deleteFuncsByScope[compute.ForwardingRule]{
			global:   gcp.GlobalForwardingRules().Delete,
			regional: gcp.ForwardingRules().Delete,
		},
		alpha: deleteFuncsByScope[alpha.ForwardingRule]{
			global:   gcp.AlphaGlobalForwardingRules().Delete,
			regional: gcp.AlphaForwardingRules().Delete,
		},
		beta: deleteFuncsByScope[beta.ForwardingRule]{
			global:   gcp.BetaGlobalForwardingRules().Delete,
			regional: gcp.BetaForwardingRules().Delete,
		},
	}
}
