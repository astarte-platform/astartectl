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

var astarteDevelopmentProfile010 AstarteClusterProfile = AstarteClusterProfile{
	Name:        "development",
	Description: "Mid-sized profile, meant for development and test Clusters. Deterministic and reasonable allocations, internal Cassandra, no replication. Safest way to bring up a reliable, well-performing development Cluster.",
	Requirements: AstarteProfileRequirements{
		CPUAllocation:    7000,
		MemoryAllocation: 20 * 1024 * 1024 * 1024,
	},
	Compatibility:      AstarteProfileCompatibility{},
	DefaultSpec:        Astartev1alpha1DeploymentSpec{},
	CustomizableFields: []AstarteProfileCustomizableField{},
}

func init() {
	astarteDevelopmentProfile010.Compatibility.MaxAstarteVersion, _ = semver.NewVersion("0.10.99")

	// Let components basic only
	astarteDevelopmentProfile010.DefaultSpec.Components.Resources.Requests.CPU = "3000m"
	astarteDevelopmentProfile010.DefaultSpec.Components.Resources.Requests.Memory = "6144M"
	astarteDevelopmentProfile010.DefaultSpec.Components.Resources.Limits.CPU = "5000m"
	astarteDevelopmentProfile010.DefaultSpec.Components.Resources.Limits.Memory = "10240M"

	// Very tiny Cassandra installation
	astarteDevelopmentProfile010.DefaultSpec.Cassandra.Deploy = true
	astarteDevelopmentProfile010.DefaultSpec.Cassandra.MaxHeapSize = "2048M"
	astarteDevelopmentProfile010.DefaultSpec.Cassandra.HeapNewSize = "256M"
	astarteDevelopmentProfile010.DefaultSpec.Cassandra.Resources.Requests.CPU = "1000m"
	astarteDevelopmentProfile010.DefaultSpec.Cassandra.Resources.Requests.Memory = "2048M"
	astarteDevelopmentProfile010.DefaultSpec.Cassandra.Resources.Limits.CPU = "2000m"
	astarteDevelopmentProfile010.DefaultSpec.Cassandra.Resources.Limits.Memory = "4096M"
	astarteDevelopmentProfile010.DefaultSpec.Cassandra.Storage.Size = "30Gi"

	// Minimal CFSSL installation
	astarteDevelopmentProfile010.DefaultSpec.Cfssl.Deploy = true
	astarteDevelopmentProfile010.DefaultSpec.Cfssl.Resources.Requests.CPU = "200m"
	astarteDevelopmentProfile010.DefaultSpec.Cfssl.Resources.Requests.Memory = "128M"
	astarteDevelopmentProfile010.DefaultSpec.Cfssl.Resources.Limits.CPU = "400m"
	astarteDevelopmentProfile010.DefaultSpec.Cfssl.Resources.Limits.Memory = "256M"
	astarteDevelopmentProfile010.DefaultSpec.Cfssl.Storage.Size = "2Gi"

	// Mid-sized RabbitMQ installation
	astarteDevelopmentProfile010.DefaultSpec.Rabbitmq.Deploy = true
	astarteDevelopmentProfile010.DefaultSpec.Rabbitmq.Resources.Requests.CPU = "1000m"
	astarteDevelopmentProfile010.DefaultSpec.Rabbitmq.Resources.Requests.Memory = "1024M"
	astarteDevelopmentProfile010.DefaultSpec.Rabbitmq.Resources.Limits.CPU = "2000m"
	astarteDevelopmentProfile010.DefaultSpec.Rabbitmq.Resources.Limits.Memory = "2048M"
	astarteDevelopmentProfile010.DefaultSpec.Rabbitmq.Storage.Size = "4Gi"

	// Minimal VerneMQ installation
	astarteDevelopmentProfile010.DefaultSpec.Vernemq.Deploy = true
	astarteDevelopmentProfile010.DefaultSpec.Vernemq.Resources.Requests.CPU = "500m"
	astarteDevelopmentProfile010.DefaultSpec.Vernemq.Resources.Requests.Memory = "1024M"
	astarteDevelopmentProfile010.DefaultSpec.Vernemq.Resources.Limits.CPU = "1000m"
	astarteDevelopmentProfile010.DefaultSpec.Vernemq.Resources.Limits.Memory = "2048M"
	astarteDevelopmentProfile010.DefaultSpec.Vernemq.Storage.Size = "4Gi"
}
