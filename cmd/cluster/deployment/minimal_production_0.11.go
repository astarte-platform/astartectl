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
	"go/types"

	"github.com/Masterminds/semver/v3"
)

var astarteMinimalProductionProfile011 AstarteClusterProfile = AstarteClusterProfile{
	Name:        "minimal-production",
	Description: "Minimal production environment for cluster with low loads. Requires an external Cassandra installation, and has no replication.",
	Requirements: AstarteProfileRequirements{
		CPUAllocation:    4 * 1000,
		MemoryAllocation: 13 * 1024 * 1024 * 1024,
		MinNodes:         1,
	},
	Compatibility: AstarteProfileCompatibility{},
	DefaultSpec:   Astartev1alpha1DeploymentSpec{},
	CustomizableFields: []AstarteProfileCustomizableField{
		// Custom Data queue size - default to 1024
		{
			Field:    "components.dataUpdaterPlant.dataQueueCount",
			Question: "Please enter the number of queues to assign to Data Updater Plant:",
			Default:  1024, AllowEmpty: false, Type: types.Int,
		},
	},
}

func init() {
	astarteMinimalProductionProfile011.Compatibility.MaxAstarteVersion, _ = semver.NewVersion("0.11.99")
	astarteMinimalProductionProfile011.Compatibility.MinAstarteVersion, _ = semver.NewVersion("0.11.0")

	// Let components burst only
	astarteMinimalProductionProfile011.DefaultSpec.Components.Resources.Requests.CPU = "1600m"
	astarteMinimalProductionProfile011.DefaultSpec.Components.Resources.Requests.Memory = "4096M"
	astarteMinimalProductionProfile011.DefaultSpec.Components.Resources.Limits.CPU = "3200m"
	astarteMinimalProductionProfile011.DefaultSpec.Components.Resources.Limits.Memory = "8192M"

	// External Cassandra installation
	astarteMinimalProductionProfile011.DefaultSpec.Cassandra.Deploy = false

	// Minimal CFSSL installation
	astarteMinimalProductionProfile011.DefaultSpec.Cfssl.Deploy = true
	astarteMinimalProductionProfile011.DefaultSpec.Cfssl.Resources.Requests.CPU = "100m"
	astarteMinimalProductionProfile011.DefaultSpec.Cfssl.Resources.Requests.Memory = "128M"
	astarteMinimalProductionProfile011.DefaultSpec.Cfssl.Resources.Limits.CPU = "200m"
	astarteMinimalProductionProfile011.DefaultSpec.Cfssl.Resources.Limits.Memory = "256M"
	astarteMinimalProductionProfile011.DefaultSpec.Cfssl.Storage.Size = "2Gi"

	// Minimal, single-replica RabbitMQ installation
	astarteMinimalProductionProfile011.DefaultSpec.Rabbitmq.Deploy = true
	astarteMinimalProductionProfile011.DefaultSpec.Rabbitmq.Resources.Requests.CPU = "600m"
	astarteMinimalProductionProfile011.DefaultSpec.Rabbitmq.Resources.Requests.Memory = "1024M"
	astarteMinimalProductionProfile011.DefaultSpec.Rabbitmq.Resources.Limits.CPU = "1000m"
	astarteMinimalProductionProfile011.DefaultSpec.Rabbitmq.Resources.Limits.Memory = "2048M"
	astarteMinimalProductionProfile011.DefaultSpec.Rabbitmq.Storage.Size = "4Gi"

	// Minimal VerneMQ installation
	astarteMinimalProductionProfile011.DefaultSpec.Vernemq.Deploy = true
	astarteMinimalProductionProfile011.DefaultSpec.Vernemq.Resources.Requests.CPU = "300m"
	astarteMinimalProductionProfile011.DefaultSpec.Vernemq.Resources.Requests.Memory = "1024M"
	astarteMinimalProductionProfile011.DefaultSpec.Vernemq.Resources.Limits.CPU = "1000m"
	astarteMinimalProductionProfile011.DefaultSpec.Vernemq.Resources.Limits.Memory = "2048M"
	astarteMinimalProductionProfile011.DefaultSpec.Vernemq.Storage.Size = "4Gi"
}
