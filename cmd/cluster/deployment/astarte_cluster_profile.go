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

// AstarteProfileCustomizableField represents a customizable field in a deployment profile
type AstarteProfileCustomizableField struct {
	Field      string          `yaml:"field"`
	Question   string          `yaml:"question"`
	Default    interface{}     `yaml:"default"`
	Type       types.BasicKind `yaml:"type"`
	AllowEmpty bool            `yaml:"allowEmpty,omitempty"`
}

// AstarteProfileCompatibility represents a compatibility range for an Astarte profile
type AstarteProfileCompatibility struct {
	MinAstarteVersion *semver.Version `yaml:"minAstarteVersion,omitempty"`
	MaxAstarteVersion *semver.Version `yaml:"maxAstarteVersion,omitempty"`
}

// AstarteProfileRequirements represents the requirements for an Astarte profile
type AstarteProfileRequirements struct {
	CPUAllocation    int64 `yaml:"cpuAllocation"`
	MemoryAllocation int64 `yaml:"memoryAllocation"`
	MinNodes         int   `yaml:"minNodes,omitempty"`
	MaxNodes         int   `yaml:"maxNodes,omitempty"`
}

// AstarteClusterProfile represents a deployment profile for an Astarte Cluster
type AstarteClusterProfile struct {
	Name               string                            `yaml:"name"`
	Description        string                            `yaml:"description"`
	Compatibility      AstarteProfileCompatibility       `yaml:"compatibility"`
	Requirements       AstarteProfileRequirements        `yaml:"requirements"`
	DefaultSpec        AstarteDeploymentSpec             `yaml:"defaultSpec"`
	CustomizableFields []AstarteProfileCustomizableField `yaml:"customizableFields"`
}

// IsValid returns whether the AstarteClusterProfile is valid or not.
func (p *AstarteClusterProfile) IsValid() bool {
	for _, s := range GetAllBuiltinAstarteClusterProfiles() {
		if s.Name == p.Name {
			return true
		}
	}

	return false
}

// GetProfilesForVersionAndRequirements gets all profiles compatible with given version and requirements
func GetProfilesForVersionAndRequirements(version *semver.Version, requirements AstarteProfileRequirements) map[string]AstarteClusterProfile {
	ret := map[string]AstarteClusterProfile{}
	for _, v := range GetAllBuiltinAstarteClusterProfiles() {
		if v.Compatibility.MinAstarteVersion != nil {
			if version.LessThan(v.Compatibility.MinAstarteVersion) {
				continue
			}
		}
		if v.Compatibility.MaxAstarteVersion != nil {
			if version.GreaterThan(v.Compatibility.MaxAstarteVersion) {
				continue
			}
		}
		if requirements.CPUAllocation < v.Requirements.CPUAllocation {
			continue
		}
		if requirements.MemoryAllocation < v.Requirements.MemoryAllocation {
			continue
		}
		if requirements.MinNodes < v.Requirements.MinNodes && v.Requirements.MinNodes > 0 {
			continue
		}
		if requirements.MaxNodes > v.Requirements.MaxNodes && v.Requirements.MaxNodes > 0 {
			continue
		}
		ret[v.Name] = v
	}

	return ret
}

// GetMatchingProfile returns a profile which matches name and version requirements, or an invalid profile.
func GetMatchingProfile(name string, version *semver.Version) AstarteClusterProfile {
	for _, profile := range GetAllBuiltinAstarteClusterProfiles() {
		if profile.Name == name {
			// Check version constraints
			if profile.Compatibility.MinAstarteVersion != nil {
				if profile.Compatibility.MinAstarteVersion.GreaterThan(version) {
					continue
				}
			}
			if profile.Compatibility.MaxAstarteVersion != nil {
				if profile.Compatibility.MaxAstarteVersion.LessThan(version) {
					continue
				}
			}

			// It's a match!
			return profile
		}
	}

	return AstarteClusterProfile{}
}
