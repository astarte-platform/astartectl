// Copyright © 2019 - 23 SECO Mind Srl
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
		astarteBasicProfile010,
		// 0.11 profiles
		astarteBasicProfile011,
		astarteBurstProfile011,
		// 1.0 profiles (good for 1.0 for now)
		astarteBasicProfile10,
		astarteBurstProfile10,
		// 1.1 profiles
		astarteBasicProfile11,
		astarteBurstProfile11,
	}
}
