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
	callURL.Path = path.Join(callURL.Path, "/v1/"+realm+"/interfaces")
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
