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

// GetAllBuiltinAstarteClusterProfiles returns all the bundled Cluster profiles.
func GetAllBuiltinAstarteClusterProfiles() []AstarteClusterProfile {
	return []AstarteClusterProfile{
		// 0.10 profiles
		astarteBurstProfile010,
		astarteBasicProfile010,
		astarteMinimalProductionProfile010,
		astarteDevelopmentProfile010,
		// 0.11 profiles
		astarteBurstProfile011,
		astarteBasicProfile011,
		astarteMinimalProductionProfile011,
		astarteDevelopmentProfile011,
	}
}
