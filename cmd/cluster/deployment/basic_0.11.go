// Copyright © 2019 Ispirata Srl
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package deployment

import (
	"github.com/Masterminds/semver/v3"
)

var astarteBasicProfile011 AstarteClusterProfile = AstarteClusterProfile{
	Name:        "basic",
	Description: "Basic profile for test Clusters. Deterministic allocations, all Astarte Pods work on basics.",
	Requirements: AstarteProfileRequirements{
		CPUAllocation:    3500,
		MemoryAllocation: 7 * 1024 * 1024 * 1024,
	},
	Compatibility:      AstarteProfileCompatibility{},
	DefaultSpec:        AstarteDeploymentSpec{},
	CustomizableFields: []AstarteProfileCustomizableField{},
}

func init() {
	astarteBasicProfile011.Compatibility.MaxAstarteVersion, _ = semver.NewVersion("0.11.99")
	astarteBasicProfile011.Compatibility.MinAstarteVersion, _ = semver.NewVersion("0.11.0")

	// Let components basic only
	astarteBasicProfile011.DefaultSpec.Components.Resources.Requests.CPU = "1200m"
	astarteBasicProfile011.DefaultSpec.Components.Resources.Requests.Memory = "2048M"
	astarteBasicProfile011.DefaultSpec.Components.Resources.Limits.CPU = "3000m"
	astarteBasicProfile011.DefaultSpec.Components.Resources.Limits.Memory = "3072M"

	// Queue size to a minimum, decent amount
	astarteBasicProfile011.DefaultSpec.Components.DataUpdaterPlant.DataQueueCount = 128

	// Very tiny Cassandra installation
	astarteBasicProfile011.DefaultSpec.Cassandra.Deploy = true
	astarteBasicProfile011.DefaultSpec.Cassandra.MaxHeapSize = "1024M"
	astarteBasicProfile011.DefaultSpec.Cassandra.HeapNewSize = "256M"
	astarteBasicProfile011.DefaultSpec.Cassandra.Resources.Requests.CPU = "1000m"
	astarteBasicProfile011.DefaultSpec.Cassandra.Resources.Requests.Memory = "1024M"
	astarteBasicProfile011.DefaultSpec.Cassandra.Resources.Limits.CPU = "2000m"
	astarteBasicProfile011.DefaultSpec.Cassandra.Resources.Limits.Memory = "2048M"
	astarteBasicProfile011.DefaultSpec.Cassandra.Storage.Size = "30Gi"

	// Minimal CFSSL installation
	astarteBasicProfile011.DefaultSpec.Cfssl.Deploy = true
	astarteBasicProfile011.DefaultSpec.Cfssl.Resources.Requests.CPU = "100m"
	astarteBasicProfile011.DefaultSpec.Cfssl.Resources.Requests.Memory = "128M"
	astarteBasicProfile011.DefaultSpec.Cfssl.Resources.Limits.CPU = "200m"
	astarteBasicProfile011.DefaultSpec.Cfssl.Resources.Limits.Memory = "256M"
	astarteBasicProfile011.DefaultSpec.Cfssl.Storage.Size = "2Gi"

	// Minimal RabbitMQ installation
	astarteBasicProfile011.DefaultSpec.Rabbitmq.Deploy = true
	astarteBasicProfile011.DefaultSpec.Rabbitmq.Resources.Requests.CPU = "300m"
	astarteBasicProfile011.DefaultSpec.Rabbitmq.Resources.Requests.Memory = "512M"
	astarteBasicProfile011.DefaultSpec.Rabbitmq.Resources.Limits.CPU = "1000m"
	astarteBasicProfile011.DefaultSpec.Rabbitmq.Resources.Limits.Memory = "1024M"
	astarteBasicProfile011.DefaultSpec.Rabbitmq.Storage.Size = "4Gi"

	// Minimal VerneMQ installation
	astarteBasicProfile011.DefaultSpec.Vernemq.Deploy = true
	astarteBasicProfile011.DefaultSpec.Vernemq.Resources.Requests.CPU = "200m"
	astarteBasicProfile011.DefaultSpec.Vernemq.Resources.Requests.Memory = "1024M"
	astarteBasicProfile011.DefaultSpec.Vernemq.Resources.Limits.CPU = "1000m"
	astarteBasicProfile011.DefaultSpec.Vernemq.Resources.Limits.Memory = "1024M"
	astarteBasicProfile011.DefaultSpec.Vernemq.Storage.Size = "4Gi"
}
