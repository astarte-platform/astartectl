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

package client

import (
	"bytes"
	"encoding/json"
	"time"
)

// ReplicationClass represents different Replication Strategies for a Realm.
type ReplicationClass int

const (
	// SimpleStrategy represents a Simple Replication Class, with a single Replication Factor
	SimpleStrategy ReplicationClass = iota
	// NetworkTopologyStrategy represents a Replication spread across DataCenters, with individual Replication Factors.
	NetworkTopologyStrategy
)

func (s ReplicationClass) String() string {
	return toString[s]
}

var toString = map[ReplicationClass]string{
	SimpleStrategy:          "SimpleStrategy",
	NetworkTopologyStrategy: "NetworkTopologyStrategy",
}

var toID = map[string]ReplicationClass{
	"SimpleStrategy":          SimpleStrategy,
	"NetworkTopologyStrategy": NetworkTopologyStrategy,
}

// MarshalJSON marshals the enum as a quoted json string
func (s ReplicationClass) MarshalJSON() ([]byte, error) {
	buffer := bytes.NewBufferString(`"`)
	buffer.WriteString(toString[s])
	buffer.WriteString(`"`)
	return buffer.Bytes(), nil
}

// UnmarshalJSON unmashals a quoted json string to the enum value
func (s *ReplicationClass) UnmarshalJSON(b []byte) error {
	var j string
	err := json.Unmarshal(b, &j)
	if err != nil {
		return err
	}
	// Note that if the string cannot be found then it will be set to the zero value, 'SimpleStrategy' in this case.
	*s = toID[j]
	return nil
}

// Not exported as it's for internal use
type deviceRegistrationResponse struct {
	CredentialsSecret string `json:"credentials_secret"`
}

// RealmDetails represents details of a single Realm
type RealmDetails struct {
	Name                         string           `json:"realm_name"`
	JwtPublicKeyPEM              string           `json:"jwt_public_key_pem"`
	ReplicationClass             ReplicationClass `json:"replication_class,omitempty"`
	ReplicationFactor            int              `json:"replication_factor,omitempty"`
	DatacenterReplicationFactors map[string]int   `json:"datacenter_replication_factors,omitempty"`
}

// DeviceInterfaceIntrospection represents a single entry in a Device Introspection array retrieved
// from DeviceDetails
type DeviceInterfaceIntrospection struct {
	Major int `json:"major"`
	Minor int `json:"minor"`
}

// DeviceDetails maps to the JSON object returned by a Device Details call to AppEngine API
type DeviceDetails struct {
	TotalReceivedMessages    int                                     `json:"total_received_msgs"`
	TotalReceivedBytes       int                                     `json:"total_received_bytes"`
	LastSeenIP               string                                  `json:"last_seen_ip"`
	LastDisconnection        time.Time                               `json:"last_disconnection"`
	LastCredentialsRequestIP string                                  `json:"last_credentials_request_ip"`
	LastConnection           time.Time                               `json:"last_connection"`
	DeviceID                 string                                  `json:"id"`
	FirstRegistration        time.Time                               `json:"first_registration"`
	FirstCredentialsRequest  time.Time                               `json:"first_credentials_request"`
	Connected                bool                                    `json:"connected"`
	Introspection            map[string]DeviceInterfaceIntrospection `json:"introspection"`
	Aliases                  map[string]string                       `json:"aliases"`
}

// DatastreamValue represent one single Datastream Value
type DatastreamValue struct {
	Value              interface{} `json:"value"`
	Timestamp          time.Time   `json:"timestamp"`
	ReceptionTimestamp time.Time   `json:"reception_timestamp"`
}

// DatastreamAggregateValue represent one single Datastream Value for an Aggregate
type DatastreamAggregateValue struct {
	Values             map[string]interface{} `json:"value"`
	Timestamp          time.Time              `json:"timestamp"`
	ReceptionTimestamp time.Time              `json:"reception_timestamp"`
}
