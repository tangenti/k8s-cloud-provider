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

package meta

import (
	"reflect"

	ga "google.golang.org/api/networkservices/v1"
	beta "google.golang.org/api/networkservices/v1beta1"
)

func init() {
	AllServices = append(AllServices, NetworkServices...)
}

const NetworkServicesAPIGroup = "networkservices"

var NetworkServices = []*ServiceInfo{
	{
		Object:      "TcpRoute",
		Service:     "TcpRoutes",
		Resource:    "tcpRoutes",
		APIGroup:    NetworkServicesAPIGroup,
		version:     VersionGA,
		keyType:     Global,
		serviceType: reflect.TypeOf(&ga.ProjectsLocationsTcpRoutesService{}),
		additionalMethods: []string{
			"Patch",
		},
	},
	{
		Object:      "TcpRoute",
		Service:     "TcpRoutes",
		Resource:    "tcpRoutes",
		APIGroup:    NetworkServicesAPIGroup,
		version:     VersionBeta,
		keyType:     Global,
		serviceType: reflect.TypeOf(&beta.ProjectsLocationsTcpRoutesService{}),
		additionalMethods: []string{
			"Patch",
		},
	},
	{
		Object:      "Mesh",
		Service:     "Meshes",
		Resource:    "meshes",
		APIGroup:    NetworkServicesAPIGroup,
		version:     VersionGA,
		keyType:     Global,
		serviceType: reflect.TypeOf(&ga.ProjectsLocationsMeshesService{}),
		additionalMethods: []string{
			"Patch",
		},
	},
	{
		Object:      "Mesh",
		Service:     "Meshes",
		Resource:    "meshes",
		APIGroup:    NetworkServicesAPIGroup,
		version:     VersionBeta,
		keyType:     Global,
		serviceType: reflect.TypeOf(&beta.ProjectsLocationsMeshesService{}),
		additionalMethods: []string{
			"Patch",
		},
	},
}
