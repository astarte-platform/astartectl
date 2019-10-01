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
	"fmt"
	"net/url"
	"path"
)

// RealmManagementService is the API Client for RealmManagement API
type RealmManagementService struct {
	client             *Client
	realmManagementURL *url.URL
}

// ListInterfaces returns all interfaces in a Realm.
func (s *RealmManagementService) ListInterfaces(realm string, token string) ([]string, error) {
	callURL, _ := url.Parse(s.realmManagementURL.String())
	callURL.Path = path.Join(callURL.Path, fmt.Sprintf("/v1/%s/interfaces", realm))
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

// ListInterfaceMajorVersions returns all available major versions for a given Interface in a Realm.
func (s *RealmManagementService) ListInterfaceMajorVersions(realm string, interfaceName string, token string) ([]int, error) {
	callURL, _ := url.Parse(s.realmManagementURL.String())
	callURL.Path = path.Join(callURL.Path, fmt.Sprintf("/v1/%s/interfaces/%s", realm, interfaceName))
	decoder, err := s.client.genericJSONDataAPIGET(callURL.String(), token, 200)
	if err != nil {
		return nil, err
	}
	var responseBody struct {
		Data []int `json:"data"`
	}
	err = decoder.Decode(&responseBody)
	if err != nil {
		return nil, err
	}

	return responseBody.Data, nil
}

// GetInterface returns an interface, identified by a Major version, in a Realm
func (s *RealmManagementService) GetInterface(realm string, interfaceName string, interfaceMajor int, token string) (map[string]interface{}, error) {
	callURL, _ := url.Parse(s.realmManagementURL.String())
	callURL.Path = path.Join(callURL.Path, fmt.Sprintf("/v1/%s/interfaces/%s/%v", realm, interfaceName, interfaceMajor))
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

// InstallInterface installs a new major version of an Interface into the Realm
func (s *RealmManagementService) InstallInterface(realm string, interfacePayload interface{}, token string) error {
	callURL, _ := url.Parse(s.realmManagementURL.String())
	callURL.Path = path.Join(callURL.Path, fmt.Sprintf("/v1/%s/interfaces", realm))
	return s.client.genericJSONDataAPIPost(callURL.String(), interfacePayload, token, 201)
}

// DeleteInterface deletes a draft Interface from the Realm
func (s *RealmManagementService) DeleteInterface(realm string, interfaceName string, interfaceMajor int, token string) error {
	callURL, _ := url.Parse(s.realmManagementURL.String())
	callURL.Path = path.Join(callURL.Path, fmt.Sprintf("/v1/%s/interfaces/%s/%v", realm, interfaceName, interfaceMajor))
	return s.client.genericJSONDataAPIDelete(callURL.String(), token, 204)
}

// UpdateInterface updates an existing major version of an Interface to a new minor.
func (s *RealmManagementService) UpdateInterface(realm string, interfaceName string, interfaceMajor int, interfacePayload interface{}, token string) error {
	callURL, _ := url.Parse(s.realmManagementURL.String())
	callURL.Path = path.Join(callURL.Path, fmt.Sprintf("/v1/%s/interfaces/%s/%v", realm, interfaceName, interfaceMajor))
	return s.client.genericJSONDataAPIPut(callURL.String(), interfacePayload, token, 201)
}
