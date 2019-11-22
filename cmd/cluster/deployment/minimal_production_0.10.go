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

import "github.com/Masterminds/semver/v3"

var astarteMinimalProductionProfile010 AstarteClusterProfile = AstarteClusterProfile{
	Name:        "minimal-production",
	Description: "Minimal production environment for cluster with low loads. Requires an external Cassandra installation, and has no replication.",
	Requirements: AstarteProfileRequirements{
		CPUAllocation:    4 * 1000,
		MemoryAllocation: 13 * 1024 * 1024 * 1024,
		MinNodes:         1,
	},
	Compatibility:      AstarteProfileCompatibility{},
	DefaultSpec:        Astartev1alpha1DeploymentSpec{},
	CustomizableFields: []AstarteProfileCustomizableField{},
}

func init() {
	astarteMinimalProductionProfile010.Compatibility.MaxAstarteVersion, _ = semver.NewVersion("0.10.99")

	// Let components burst only
	astarteMinimalProductionProfile010.DefaultSpec.Components.Resources.Requests.CPU = "1600m"
	astarteMinimalProductionProfile010.DefaultSpec.Components.Resources.Requests.Memory = "4096M"
	astarteMinimalProductionProfile010.DefaultSpec.Components.Resources.Limits.CPU = "3200m"
	astarteMinimalProductionProfile010.DefaultSpec.Components.Resources.Limits.Memory = "8192M"

	// External Cassandra installation
	astarteMinimalProductionProfile010.DefaultSpec.Cassandra.Deploy = false

	// Minimal CFSSL installation
	astarteMinimalProductionProfile010.DefaultSpec.Cfssl.Deploy = true
	astarteMinimalProductionProfile010.DefaultSpec.Cfssl.Resources.Requests.CPU = "100m"
	astarteMinimalProductionProfile010.DefaultSpec.Cfssl.Resources.Requests.Memory = "128M"
	astarteMinimalProductionProfile010.DefaultSpec.Cfssl.Resources.Limits.CPU = "200m"
	astarteMinimalProductionProfile010.DefaultSpec.Cfssl.Resources.Limits.Memory = "256M"
	astarteMinimalProductionProfile010.DefaultSpec.Cfssl.Storage.Size = "2Gi"

	// Minimal, single-replica RabbitMQ installation
	astarteMinimalProductionProfile010.DefaultSpec.Rabbitmq.Deploy = true
	astarteMinimalProductionProfile010.DefaultSpec.Rabbitmq.Resources.Requests.CPU = "600m"
	astarteMinimalProductionProfile010.DefaultSpec.Rabbitmq.Resources.Requests.Memory = "1024M"
	astarteMinimalProductionProfile010.DefaultSpec.Rabbitmq.Resources.Limits.CPU = "1000m"
	astarteMinimalProductionProfile010.DefaultSpec.Rabbitmq.Resources.Limits.Memory = "2048M"
	astarteMinimalProductionProfile010.DefaultSpec.Rabbitmq.Storage.Size = "4Gi"

	// Minimal VerneMQ installation
	astarteMinimalProductionProfile010.DefaultSpec.Vernemq.Deploy = true
	astarteMinimalProductionProfile010.DefaultSpec.Vernemq.Resources.Requests.CPU = "300m"
	astarteMinimalProductionProfile010.DefaultSpec.Vernemq.Resources.Requests.Memory = "1024M"
	astarteMinimalProductionProfile010.DefaultSpec.Vernemq.Resources.Limits.CPU = "1000m"
	astarteMinimalProductionProfile010.DefaultSpec.Vernemq.Resources.Limits.Memory = "2048M"
	astarteMinimalProductionProfile010.DefaultSpec.Vernemq.Storage.Size = "4Gi"
}
