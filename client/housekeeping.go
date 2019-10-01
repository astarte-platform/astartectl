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
	"errors"
	"fmt"
	"net/url"
	"path"
)

// HousekeepingService is the API Client for Housekeeping API
type HousekeepingService struct {
	client          *Client
	housekeepingURL *url.URL
}

// ListRealms returns all realms in the cluster.
func (s *HousekeepingService) ListRealms(token string) ([]string, error) {
	callURL, _ := url.Parse(s.housekeepingURL.String())
	callURL.Path = path.Join(callURL.Path, "/v1/realms")
	decoder, err := s.client.genericJSONDataAPIGET(callURL.String(), token, 200)
	if err != nil {
		return nil, err
	}
	var responseBody struct {
		Data []string `json:"data"`
	}
	err = decoder.Decode(&responseBody)
	if err != nil {
		return nil, err
	}

	return responseBody.Data, nil
}

// GetRealm returns data about a single Realm.
func (s *HousekeepingService) GetRealm(realm string, token string) (map[string]interface{}, error) {
	callURL, _ := url.Parse(s.housekeepingURL.String())
	callURL.Path = path.Join(callURL.Path, fmt.Sprintf("/v1/realms/%s", realm))
	decoder, err := s.client.genericJSONDataAPIGET(callURL.String(), token, 200)
	if err != nil {
		return nil, err
	}
	var responseBody struct {
		Data map[string]interface{} `json:"data"`
	}
	err = decoder.Decode(&responseBody)
	if err != nil {
		return nil, err
	}

	return responseBody.Data, nil
}

// CreateRealm creates a new Realm in the Cluster with default parameters.
func (s *HousekeepingService) CreateRealm(realm string, publicKeyString string, token string) error {
	return s.createRealmInternal(realm, publicKeyString, 0, nil, token)
}

// CreateRealmWithReplicationFactor creates a new Realm in the Cluster with a custom Replication Factor.
// The replication factor must always be > 0.
func (s *HousekeepingService) CreateRealmWithReplicationFactor(realm string, publicKeyString string,
	replicationFactor int, token string) error {
	if replicationFactor <= 0 {
		return errors.New("Replication factor should be > 0")
	}
	return s.createRealmInternal(realm, publicKeyString, replicationFactor, nil, token)
}

// CreateRealmWithDatacenterReplication creates a new Realm in the Cluster with a custom,
// per-datacenter Replication Factor. Both replicationClass and datacenterReplicationFactors must be provided.
func (s *HousekeepingService) CreateRealmWithDatacenterReplication(realm string, publicKeyString string,
	datacenterReplicationFactors map[string]int, token string) error {
	return s.createRealmInternal(realm, publicKeyString, 0, datacenterReplicationFactors, token)
}

func (s *HousekeepingService) createRealmInternal(realm string, publicKeyString string, replicationFactor int,
	datacenterReplicationFactors map[string]int, token string) error {
	callURL, _ := url.Parse(s.housekeepingURL.String())
	callURL.Path = path.Join(callURL.Path, "/v1/realms")

	requestBody := map[string]interface{}{
		"realm_name":         realm,
		"jwt_public_key_pem": publicKeyString,
	}

	if replicationFactor > 0 {
		requestBody["replication_factor"] = replicationFactor
	} else if datacenterReplicationFactors != nil {
		requestBody["replication_class"] = "NetworkTopologyStrategy"
		requestBody["datacenter_replication_factors"] = datacenterReplicationFactors
	}

	return s.client.genericJSONDataAPIPost(callURL.String(), requestBody, token, 201)
}
