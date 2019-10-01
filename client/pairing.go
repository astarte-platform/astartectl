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

// PairingService is the API Client for Pairing API
type PairingService struct {
	client     *Client
	pairingURL *url.URL
}

// RegisterDevice registers a new device into the Realm.
// Returns the Credential Secret of the Device when successful.
// TODO: add support for initial_introspection
func (s *PairingService) RegisterDevice(realm string, deviceID string, token string) (string, error) {
	callURL, _ := url.Parse(s.pairingURL.String())
	callURL.Path = path.Join(callURL.Path, fmt.Sprintf("/v1/%s/agent/devices", realm))

	var requestBody struct {
		HwID string `json:"hw_id"`
	}
	requestBody.HwID = deviceID

	decoder, err := s.client.genericJSONDataAPIPostWithResponse(callURL.String(), requestBody, token, 201)
	if err != nil {
		return "", err
	}

	// Decode the reply
	var responseBody struct {
		Data deviceRegistrationResponse `json:"data"`
	}
	err = decoder.Decode(&responseBody)
	if err != nil {
		return "", err
	}

	return responseBody.Data.CredentialsSecret, nil
}
