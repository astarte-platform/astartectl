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

	"github.com/astarte-platform/astarte-go/astarteservices"
	"github.com/astarte-platform/astarte-go/client"
	"github.com/spf13/viper"
)

// APICommandSetup is a helper for setting up a generic command using Astarte API.
// individualURLs must contain the service->variable association.
func APICommandSetup(individualURLVariables map[astarteservices.AstarteService]string, keyVariable, keyFileVariable string) (*client.Client, error) {
	var clientConfig = []client.Option{}

	httpConfig := setupHTTP()
	clientConfig = append(clientConfig, httpConfig...)

	authConfig, err := setupAuth(keyVariable, keyFileVariable)
	if err != nil {
		return nil, err
	}
	clientConfig = append(clientConfig, authConfig...)

	URLConfig, err := setupURLs(individualURLVariables)
	if err != nil {
		return nil, err
	}
	clientConfig = append(clientConfig, URLConfig...)

	astarteAPIClient, err := client.New(clientConfig...)
	if err != nil {
		return nil, err
	}

	return astarteAPIClient, nil
}

func setupHTTP() []client.Option {
	var ret = []client.Option{}
	ignoreSSLErrors := viper.GetBool("ignore-ssl-errors")
	if ignoreSSLErrors {
		httpClient := &http.Client{
			Timeout: time.Second * 30,
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: ignoreSSLErrors,
				},
			},
		}
		ret = append(ret, client.WithHTTPClient(httpClient))
	}
	return ret
}

func setupAuth(keyVariable, keyFileVariable string) ([]client.Option, error) {
	var ret = []client.Option{}

	// Setup auth
	privateKeyFile := viper.GetString(keyFileVariable)
	privateKey := viper.GetString(keyVariable)
	explicitToken := viper.GetString("token")
	if privateKey == "" && privateKeyFile == "" && explicitToken == "" {
		return nil, fmt.Errorf("%s or token is required", strings.Replace(keyFileVariable, ".", "-", -1))
	}
	if explicitToken == "" {
		// 1 minute TTL is more than enough for our purposes
		if privateKeyFile != "" {
			ret = append(ret, client.WithPrivateKey(privateKeyFile), client.WithExpiry(60))
		} else {
			decoded, err := base64.StdEncoding.DecodeString(privateKey)
			if err != nil {
				return nil, err
			}
			ret = append(ret, client.WithPrivateKey(decoded), client.WithExpiry(60))
		}
	} else {
		ret = append(ret, client.WithJWT(explicitToken))
	}
	return ret, nil
}

func setupURLs(individualURLVariables map[astarteservices.AstarteService]string) ([]client.Option, error) {
	var ret = []client.Option{}

	astarteURL := viper.GetString("url")
	individualURLs := map[astarteservices.AstarteService]string{}
	for k, v := range individualURLVariables {
		urlOverride := viper.GetString(v)
		if urlOverride != "" {
			individualURLs[k] = urlOverride
		}
	}

	if len(individualURLs) > 0 {
		ret = append(ret, setupIndividualURLs(individualURLs)...)
	} else if astarteURL != "" {
		ret = append(ret, client.WithBaseURL(astarteURL))
	} else {
		return nil, errors.New("Either astarte-url or an individual API URL have to be specified")
	}
	return ret, nil
}

func setupIndividualURLs(individualURLs map[astarteservices.AstarteService]string) []client.Option {
	var ret = []client.Option{}

	// Golang I hate you
	if individualURLs[astarteservices.Housekeeping] != "" {
		ret = append(ret, client.WithHousekeepingURL(individualURLs[astarteservices.Housekeeping]))
	}
	if individualURLs[astarteservices.AppEngine] != "" {
		ret = append(ret, client.WithAppEngineURL(individualURLs[astarteservices.AppEngine]))
	}
	if individualURLs[astarteservices.Pairing] != "" {
		ret = append(ret, client.WithPairingURL(individualURLs[astarteservices.Pairing]))
	}
	if individualURLs[astarteservices.RealmManagement] != "" {
		ret = append(ret, client.WithRealmManagementURL(individualURLs[astarteservices.RealmManagement]))
	}
	return ret
}
