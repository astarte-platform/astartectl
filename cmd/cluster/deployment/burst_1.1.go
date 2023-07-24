// Copyright © 2023 SECO Mind Srl
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

var astarteBurstProfile11 AstarteClusterProfile = AstarteClusterProfile{
	Name:        "burst",
	Description: "Burst profile for CI Clusters. No deterministic allocations, all Astarte Pods work on bursts.",
	Requirements: AstarteProfileRequirements{
		CPUAllocation:    2 * 1000,
		MemoryAllocation: 5 * 1024 * 1024 * 1024,
	},
	Compatibility:      AstarteProfileCompatibility{},
	DefaultSpec:        AstarteDeploymentSpec{},
	CustomizableFields: []AstarteProfileCustomizableField{},
}

func init() {
	astarteBurstProfile11.Compatibility.MaxAstarteVersion, _ = semver.NewVersion("1.1.99")
	astarteBurstProfile11.Compatibility.MinAstarteVersion, _ = semver.NewVersion("1.1.0")

	// Let components burst only
	astarteBurstProfile11.DefaultSpec.Components.Resources.Requests.CPU = "0m"
	astarteBurstProfile11.DefaultSpec.Components.Resources.Requests.Memory = "2048M"
	astarteBurstProfile11.DefaultSpec.Components.Resources.Limits.CPU = "0m"
	astarteBurstProfile11.DefaultSpec.Components.Resources.Limits.Memory = "3072M"

	// Queue size to a minimum, decent amount
	astarteBurstProfile11.DefaultSpec.Components.DataUpdaterPlant.DataQueueCount = 128

	// Very tiny Cassandra installation
	astarteBurstProfile11.DefaultSpec.Cassandra.Deploy = true
	astarteBurstProfile11.DefaultSpec.Cassandra.MaxHeapSize = "512M"
	astarteBurstProfile11.DefaultSpec.Cassandra.HeapNewSize = "256M"
	astarteBurstProfile11.DefaultSpec.Cassandra.Resources.Requests.CPU = "500m"
	astarteBurstProfile11.DefaultSpec.Cassandra.Resources.Requests.Memory = "1024M"
	astarteBurstProfile11.DefaultSpec.Cassandra.Resources.Limits.CPU = "1000m"
	astarteBurstProfile11.DefaultSpec.Cassandra.Resources.Limits.Memory = "2048M"
	astarteBurstProfile11.DefaultSpec.Cassandra.Storage.Size = "10Gi"

	// Minimal CFSSL installation
	astarteBurstProfile11.DefaultSpec.Cfssl.Deploy = true
	astarteBurstProfile11.DefaultSpec.Cfssl.Resources.Requests.CPU = "0m"
	astarteBurstProfile11.DefaultSpec.Cfssl.Resources.Requests.Memory = "128M"
	astarteBurstProfile11.DefaultSpec.Cfssl.Resources.Limits.CPU = "0m"
	astarteBurstProfile11.DefaultSpec.Cfssl.Resources.Limits.Memory = "128M"
	astarteBurstProfile11.DefaultSpec.Cfssl.Storage.Size = "2Gi"

	// Minimal RabbitMQ installation
	astarteBurstProfile11.DefaultSpec.Rabbitmq.Deploy = true
	astarteBurstProfile11.DefaultSpec.Rabbitmq.Resources.Requests.CPU = "200m"
	astarteBurstProfile11.DefaultSpec.Rabbitmq.Resources.Requests.Memory = "256M"
	astarteBurstProfile11.DefaultSpec.Rabbitmq.Resources.Limits.CPU = "1000m"
	astarteBurstProfile11.DefaultSpec.Rabbitmq.Resources.Limits.Memory = "256M"
	astarteBurstProfile11.DefaultSpec.Rabbitmq.Storage.Size = "4Gi"

	// Minimal VerneMQ installation
	astarteBurstProfile11.DefaultSpec.Vernemq.Deploy = true
	astarteBurstProfile11.DefaultSpec.Vernemq.Resources.Requests.CPU = "0m"
	astarteBurstProfile11.DefaultSpec.Vernemq.Resources.Requests.Memory = "256M"
	astarteBurstProfile11.DefaultSpec.Vernemq.Resources.Limits.CPU = "0m"
	astarteBurstProfile11.DefaultSpec.Vernemq.Resources.Limits.Memory = "256M"
	astarteBurstProfile11.DefaultSpec.Vernemq.Storage.Size = "4Gi"
}
