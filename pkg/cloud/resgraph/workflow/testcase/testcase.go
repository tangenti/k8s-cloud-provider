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

package testcase

import (
	"sort"

	"github.com/GoogleCloudPlatform/k8s-cloud-provider/pkg/cloud"
	"github.com/GoogleCloudPlatform/k8s-cloud-provider/pkg/cloud/resgraph"
	"github.com/GoogleCloudPlatform/k8s-cloud-provider/pkg/cloud/resgraph/exec"
)

type TestCase struct {
	Name        string
	Description string
	Steps       []Step
}

type Step struct {
	SetUp       func(cloud.Cloud)
	Graph       *resgraph.Graph
	WantActions []exec.Action
}

var (
	all map[string]*TestCase
)

func init() { all = map[string]*TestCase{} }

func Register(tc *TestCase) { all[tc.Name] = tc }
func Cases() []*TestCase {
	var keys []string
	for k := range all {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	var ret []*TestCase
	for _, k := range keys {
		ret = append(ret, all[k])
	}

	return ret
}

func Case(name string) *TestCase { return all[name] }
