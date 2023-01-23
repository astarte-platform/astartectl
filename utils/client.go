// Copyright © 2019-2020 Ispirata Srl
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

package utils

import (
	"crypto/tls"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/astarte-platform/astarte-go/client"
	"github.com/astarte-platform/astarte-go/misc"
	"github.com/spf13/viper"
)

// APICommandSetup is a helper for setting up a generic command using Astarte API.
// individualURLs must contain the service->variable association.
func APICommandSetup(individualURLVariables map[misc.AstarteService]string, keyVariable, keyFileVariable string) (*client.Client, error) {
	var astarteAPIClient *client.Client
	astarteURL := viper.GetString("url")
	individualURLs := map[misc.AstarteService]string{}
	for k, v := range individualURLVariables {
		urlOverride := viper.GetString(v)
		if urlOverride != "" {
			individualURLs[k] = urlOverride
		}
	}

	ignoreSSLErrors := viper.GetBool("ignore-ssl-errors")
	httpClient := &http.Client{
		Timeout: time.Second * 30,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: ignoreSSLErrors,
			},
		},
	}

	var err error
	if len(individualURLs) > 0 {
		// Use explicit URLs for Service
		astarteAPIClient, err = client.NewClientWithIndividualURLs(individualURLs, httpClient)
	} else if astarteURL != "" {
		astarteAPIClient, err = client.NewClient(astarteURL, httpClient)
	} else {
		err = errors.New("Either astarte-url or an individual API URL have to be specified")
	}

	if err != nil {
		return nil, err
	}

	privateKeyFile := viper.GetString(keyFileVariable)
	privateKey := viper.GetString(keyVariable)
	explicitToken := viper.GetString("token")
	if privateKey == "" && privateKeyFile == "" && explicitToken == "" {
		return nil, fmt.Errorf("%s or token is required", strings.Replace(keyFileVariable, ".", "-", -1))
	}

	if explicitToken == "" {
		servicesAndClaims := map[misc.AstarteService][]string{}
		for k := range individualURLVariables {
			servicesAndClaims[k] = []string{}
		}
		// 1 minute TTL is more than enough for our purposes
		if privateKeyFile != "" {
			if err := astarteAPIClient.SetTokenFromPrivateKeyFileWithClaims(privateKeyFile, servicesAndClaims, 60); err != nil {
				return nil, err
			}
		} else {
			decoded, err := base64.StdEncoding.DecodeString(privateKey)
			if err != nil {
				return nil, err
			}
			if err := astarteAPIClient.SetTokenFromPrivateKeyWithClaims(decoded, servicesAndClaims, 60); err != nil {
				return nil, err
			}
		}
	} else {
		astarteAPIClient.SetToken(explicitToken)
	}

	return astarteAPIClient, nil
}
