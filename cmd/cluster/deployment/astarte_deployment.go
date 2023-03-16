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

// AstarteGenericAPISpec represents a generic Astarte API Component in the Deployment spec
type AstarteGenericAPISpec struct {
	Replicas              int    `yaml:"replicas,omitempty"`
	DisableAuthentication bool   `yaml:"disableAuthentication,omitempty"`
	Version               string `yaml:"version,omitempty"`
	Resources             struct {
		Requests struct {
			CPU    string `yaml:"cpu"`
			Memory string `yaml:"memory"`
		} `yaml:"requests"`
		Limits struct {
			CPU    string `yaml:"cpu"`
			Memory string `yaml:"memory"`
		} `yaml:"limits"`
	} `yaml:"resources,omitempty"`
}

// AstarteGenericBackendSpec represents a generic Astarte Backend Component in the Deployment spec
type AstarteGenericBackendSpec struct {
	Replicas  int    `yaml:"replicas,omitempty"`
	Version   string `yaml:"version,omitempty"`
	Resources struct {
		Requests struct {
			CPU    string `yaml:"cpu"`
			Memory string `yaml:"memory"`
		} `yaml:"requests"`
		Limits struct {
			CPU    string `yaml:"cpu"`
			Memory string `yaml:"memory"`
		} `yaml:"limits"`
	} `yaml:"resources,omitempty"`
}

// AstarteDeploymentSpec represents the spec for Kubernetes API api.astarte-platform.org/v1alpha*/Astarte
type AstarteDeploymentSpec struct {
	Version             string      `yaml:"version"`
	ImagePullPolicy     string      `yaml:"imagePullPolicy,omitempty"`
	ImagePullSecrets    []string    `yaml:"imagePullSecrets,omitempty"`
	DistributionChannel string      `yaml:"distributionChannel,omitempty"`
	Rbac                bool        `yaml:"rbac,omitempty"`
	StorageClassName    interface{} `yaml:"storageClassName,omitempty"`
	API                 struct {
		Ssl  bool   `yaml:"ssl,omitempty"`
		Host string `yaml:"host"`
	} `yaml:"api"`
	Rabbitmq struct {
		Deploy     bool `yaml:"deploy,omitempty"`
		Connection struct {
			Host     string `yaml:"host"`
			Port     string `yaml:"port"`
			Username string `yaml:"username"`
			Password string `yaml:"password"`
			Secret   struct {
				Name        string `yaml:"name"`
				UsernameKey string `yaml:"usernameKey"`
				PasswordKey string `yaml:"passwordKey"`
			} `yaml:"secret"`
		} `yaml:"connection,omitempty"`
		Replicas     int  `yaml:"replicas,omitempty"`
		AntiAffinity bool `yaml:"antiAffinity,omitempty"`
		Storage      struct {
			Size             string      `yaml:"size"`
			ClassName        string      `yaml:"className,omitempty"`
			VolumeDefinition interface{} `yaml:"volumeDefinition,omitempty"`
		} `yaml:"storage,omitempty"`
		Resources struct {
			Requests struct {
				CPU    string `yaml:"cpu"`
				Memory string `yaml:"memory"`
			} `yaml:"requests"`
			Limits struct {
				CPU    string `yaml:"cpu"`
				Memory string `yaml:"memory"`
			} `yaml:"limits"`
		} `yaml:"resources,omitempty"`
	} `yaml:"rabbitmq"`
	Cassandra struct {
		Deploy       bool   `yaml:"deploy"`
		Version      string `yaml:"version,omitempty"`
		Nodes        string `yaml:"nodes,omitempty"`
		Replicas     int    `yaml:"replicas,omitempty"`
		AntiAffinity bool   `yaml:"antiAffinity,omitempty"`
		MaxHeapSize  string `yaml:"maxHeapSize,omitempty"`
		HeapNewSize  string `yaml:"heapNewSize,omitempty"`
		Storage      struct {
			Size             string      `yaml:"size"`
			ClassName        string      `yaml:"className,omitempty"`
			VolumeDefinition interface{} `yaml:"volumeDefinition,omitempty"`
		} `yaml:"storage,omitempty"`
		Resources struct {
			Requests struct {
				CPU    string `yaml:"cpu"`
				Memory string `yaml:"memory"`
			} `yaml:"requests"`
			Limits struct {
				CPU    string `yaml:"cpu"`
				Memory string `yaml:"memory"`
			} `yaml:"limits"`
		} `yaml:"resources,omitempty"`
	} `yaml:"cassandra"`
	Vernemq struct {
		Host                      string `yaml:"host,omitempty"`
		Port                      int    `yaml:"port,omitempty"`
		Deploy                    bool   `yaml:"deploy,omitempty"`
		Replicas                  int    `yaml:"replicas,omitempty"`
		AntiAffinity              bool   `yaml:"antiAffinity,omitempty"`
		SslListener               bool   `yaml:"sslListener,omitempty"`
		SslListenerCertSecretName string `yaml:"sslListenerCertSecretName,omitempty"`
		CaSecret                  string `yaml:"caSecret,omitempty"`
		Storage                   struct {
			Size             string      `yaml:"size"`
			ClassName        string      `yaml:"className,omitempty"`
			VolumeDefinition interface{} `yaml:"volumeDefinition,omitempty"`
		} `yaml:"storage,omitempty"`
		Resources struct {
			Requests struct {
				CPU    string `yaml:"cpu"`
				Memory string `yaml:"memory"`
			} `yaml:"requests"`
			Limits struct {
				CPU    string `yaml:"cpu"`
				Memory string `yaml:"memory"`
			} `yaml:"limits"`
		} `yaml:"resources,omitempty"`
	} `yaml:"vernemq"`
	Cfssl struct {
		Deploy            bool   `yaml:"deploy,omitempty"`
		Replicas          int    `yaml:"replicas,omitempty"`
		URL               string `yaml:"url,omitempty"`
		CaExpiry          string `yaml:"caExpiry,omitempty"`
		CertificateExpiry string `yaml:"certificateExpiry,omitempty"`
		DbConfig          struct {
			Driver     string `yaml:"driver,omitempty"`
			DataSource string `yaml:"dataSource,omitempty"`
		} `yaml:"dbConfig,omitempty"`
		Resources struct {
			Requests struct {
				CPU    string `yaml:"cpu"`
				Memory string `yaml:"memory"`
			} `yaml:"requests"`
			Limits struct {
				CPU    string `yaml:"cpu"`
				Memory string `yaml:"memory"`
			} `yaml:"limits"`
		} `yaml:"resources,omitempty"`
		Storage struct {
			Size             string      `yaml:"size"`
			ClassName        string      `yaml:"className,omitempty"`
			VolumeDefinition interface{} `yaml:"volumeDefinition,omitempty"`
		} `yaml:"storage,omitempty"`
		CsrRootCa struct {
			CN  string `yaml:"CN"`
			Key struct {
				Algo string `yaml:"algo"`
				Size int    `yaml:"size"`
			} `yaml:"key"`
			Names []struct {
				C  string `yaml:"C"`
				L  string `yaml:"L"`
				O  string `yaml:"O"`
				OU string `yaml:"OU"`
				ST string `yaml:"ST"`
			} `yaml:"names"`
			Ca struct {
				Expiry string `yaml:"expiry"`
			} `yaml:"ca"`
		} `yaml:"csrRootCa,omitempty"`
		CaRootConfig struct {
			Signing struct {
				Default struct {
					Usages       []string `yaml:"usages"`
					Expiry       string   `yaml:"expiry"`
					CaConstraint struct {
						IsCa           bool `yaml:"isCa"`
						MaxPathLen     int  `yaml:"maxPathLen"`
						MaxPathLenZero bool `yaml:"maxPathLenZero"`
					} `yaml:"caConstraint"`
				} `yaml:"default"`
			} `yaml:"signing"`
		} `yaml:"caRootConfig,omitempty"`
	} `yaml:"cfssl"`
	Components struct {
		Resources struct {
			Requests struct {
				CPU    string `yaml:"cpu"`
				Memory string `yaml:"memory"`
			} `yaml:"requests"`
			Limits struct {
				CPU    string `yaml:"cpu"`
				Memory string `yaml:"memory"`
			} `yaml:"limits"`
		} `yaml:"resources,omitempty"`
		Housekeeping struct {
			API     AstarteGenericAPISpec     `yaml:"api,omitempty"`
			Backend AstarteGenericBackendSpec `yaml:"backend,omitempty"`
		} `yaml:"housekeeping,omitempty"`
		RealmManagement struct {
			API     AstarteGenericAPISpec     `yaml:"api,omitempty"`
			Backend AstarteGenericBackendSpec `yaml:"backend,omitempty"`
		} `yaml:"realmManagement,omitempty"`
		Pairing struct {
			API     AstarteGenericAPISpec     `yaml:"api,omitempty"`
			Backend AstarteGenericBackendSpec `yaml:"backend,omitempty"`
		} `yaml:"pairing,omitempty"`
		DataUpdaterPlant struct {
			Replicas       int    `yaml:"replicas,omitempty"`
			DataQueueCount int    `yaml:"dataQueueCount,omitempty"`
			Version        string `yaml:"version,omitempty"`
			Resources      struct {
				Requests struct {
					CPU    string `yaml:"cpu"`
					Memory string `yaml:"memory"`
				} `yaml:"requests"`
				Limits struct {
					CPU    string `yaml:"cpu"`
					Memory string `yaml:"memory"`
				} `yaml:"limits"`
			} `yaml:"resources,omitempty"`
		} `yaml:"dataUpdaterPlant,omitempty"`
		AppengineAPI struct {
			Replicas              int    `yaml:"replicas,omitempty"`
			DisableAuthentication bool   `yaml:"disableAuthentication,omitempty"`
			MaxResultsLimit       int    `yaml:"maxResultsLimit,omitempty"`
			Version               string `yaml:"version,omitempty"`
			Resources             struct {
				Requests struct {
					CPU    string `yaml:"cpu"`
					Memory string `yaml:"memory"`
				} `yaml:"requests"`
				Limits struct {
					CPU    string `yaml:"cpu"`
					Memory string `yaml:"memory"`
				} `yaml:"limits"`
			} `yaml:"resources,omitempty"`
		} `yaml:"appengineApi,omitempty"`
		TriggerEngine struct {
			Deploy    bool   `yaml:"deploy,omitempty"`
			Replicas  int    `yaml:"replicas,omitempty"`
			Version   string `yaml:"version,omitempty"`
			Resources struct {
				Requests struct {
					CPU    string `yaml:"cpu"`
					Memory string `yaml:"memory"`
				} `yaml:"requests"`
				Limits struct {
					CPU    string `yaml:"cpu"`
					Memory string `yaml:"memory"`
				} `yaml:"limits"`
			} `yaml:"resources,omitempty"`
		} `yaml:"triggerEngine,omitempty"`
		Dashboard struct {
			Deploy       bool   `yaml:"deploy,omitempty"`
			Replicas     int    `yaml:"replicas,omitempty"`
			DefaultRealm string `yaml:"defaultRealm,omitempty"`
			Version      string `yaml:"version,omitempty"`
			Config       struct {
				RealmManagementAPIURL string `yaml:"realmManagementApiUrl"`
				DefaultRealm          string `yaml:"defaultRealm"`
				DefaultAuth           string `yaml:"defaultAuth"`
				Auth                  []struct {
					Type string `yaml:"type"`
				} `yaml:"auth"`
			} `yaml:"config,omitempty"`
			Resources struct {
				Requests struct {
					CPU    string `yaml:"cpu"`
					Memory string `yaml:"memory"`
				} `yaml:"requests"`
				Limits struct {
					CPU    string `yaml:"cpu"`
					Memory string `yaml:"memory"`
				} `yaml:"limits"`
			} `yaml:"resources,omitempty"`
		} `yaml:"dashboard,omitempty"`
	} `yaml:"components"`
}

// AstarteDeployment represents an Astarte Deployment object
type AstarteDeployment struct {
	APIVersion string `yaml:"apiVersion"`
	Kind       string `yaml:"kind"`
	Metadata   struct {
		Name        string            `yaml:"name"`
		Namespace   string            `yaml:"namespace"`
		Annotations map[string]string `yaml:"annotations,omitempty"`
	} `yaml:"metadata"`
	Spec AstarteDeploymentSpec `yaml:"spec"`
}

// GetBaseAstarteDeployment returns a ready to customise deployment spec for an Astarte resource.
// The resource API version (v1alpha1 or v1alpha2) is determined by the Astarte version.
func GetBaseAstarteDeployment(astarteVersion *semver.Version) AstarteDeployment {
	oldAstarteAPIVersion, _ := semver.StrictNewVersion("1.0.0")
	if astarteVersion.LessThan(oldAstarteAPIVersion) {
		return AstarteDeployment{
			APIVersion: "api.astarte-platform.org/v1alpha1",
			Kind:       "Astarte",
		}
	}
	return AstarteDeployment{
		APIVersion: "api.astarte-platform.org/v1alpha2",
		Kind:       "Astarte",
	}
}
