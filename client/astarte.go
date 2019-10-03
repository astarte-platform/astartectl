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
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"
	"time"
)

const (
	userAgent = "astarte-go"
)

// Client is the base Astarte API client. It provides access to all of Astarte's APIs.
type Client struct {
	baseURL   *url.URL
	UserAgent string

	httpClient *http.Client

	AppEngine       *AppEngineService
	Housekeeping    *HousekeepingService
	Pairing         *PairingService
	RealmManagement *RealmManagementService
}

// NewClient creates a new Astarte API client with standard URL hierarchies.
func NewClient(rawBaseURL string, httpClient *http.Client) (*Client, error) {
	if httpClient == nil {
		httpClient = &http.Client{
			Timeout: time.Second * 30,
		}
	}

	baseURL, err := url.Parse(rawBaseURL)
	if err != nil {
		return nil, err
	}

	c := &Client{httpClient: httpClient, baseURL: baseURL, UserAgent: userAgent}

	// Apparently that's how you deep-copy the URLs.
	// We're ignoring errors here as the cross-parsing cannot fail.
	appEngineURL, _ := url.Parse(baseURL.String())
	appEngineURL.Path = path.Join(appEngineURL.Path, "appengine")
	c.AppEngine = &AppEngineService{client: c, appEngineURL: appEngineURL}

	housekeepingURL, _ := url.Parse(baseURL.String())
	housekeepingURL.Path = path.Join(housekeepingURL.Path, "housekeeping")
	c.Housekeeping = &HousekeepingService{client: c, housekeepingURL: housekeepingURL}

	pairingURL, _ := url.Parse(baseURL.String())
	pairingURL.Path = path.Join(pairingURL.Path, "pairing")
	c.Pairing = &PairingService{client: c, pairingURL: pairingURL}

	realmManagementURL, _ := url.Parse(baseURL.String())
	realmManagementURL.Path = path.Join(realmManagementURL.Path, "realmmanagement")
	c.RealmManagement = &RealmManagementService{client: c, realmManagementURL: realmManagementURL}

	return c, nil
}

// NewClientWithIndividualURLs creates a new Astarte API client with custom URL hierarchies.
// If an empty string is passed as one of the URLs, the corresponding Service will not be instantiated.
func NewClientWithIndividualURLs(rawAppEngineURL string, rawHousekeepingURL string, rawPairingURL string,
	rawRealmManagementURL string, httpClient *http.Client) (*Client, error) {
	if httpClient == nil {
		httpClient = &http.Client{
			Timeout: time.Second * 30,
		}
	}

	c := &Client{httpClient: httpClient, baseURL: nil, UserAgent: userAgent}

	if rawAppEngineURL != "" {
		appEngineURL, err := url.Parse(rawAppEngineURL)
		if err != nil {
			return nil, err
		}
		c.AppEngine = &AppEngineService{client: c, appEngineURL: appEngineURL}
	}

	if rawHousekeepingURL != "" {
		housekeepingURL, err := url.Parse(rawHousekeepingURL)
		if err != nil {
			return nil, err
		}
		c.Housekeeping = &HousekeepingService{client: c, housekeepingURL: housekeepingURL}
	}

	if rawPairingURL != "" {
		pairingURL, err := url.Parse(rawPairingURL)
		if err != nil {
			return nil, err
		}
		c.Pairing = &PairingService{client: c, pairingURL: pairingURL}
	}

	if rawRealmManagementURL != "" {
		realmManagementURL, err := url.Parse(rawRealmManagementURL)
		if err != nil {
			return nil, err
		}
		c.RealmManagement = &RealmManagementService{client: c, realmManagementURL: realmManagementURL}
	}

	return c, nil
}

func errorFromJSONErrors(responseBody io.Reader) error {
	var errorBody struct {
		Errors map[string]interface{} `json:"errors"`
	}

	err := json.NewDecoder(responseBody).Decode(&errorBody)
	if err != nil {
		return err
	}

	errJSON, _ := json.MarshalIndent(&errorBody, "", "  ")
	return fmt.Errorf("%s", errJSON)
}

func (c *Client) genericJSONDataAPIGET(urlString string, authorizationToken string, expectedReturnCode int) (*json.Decoder, error) {
	req, err := http.NewRequest("GET", urlString, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", "Bearer "+authorizationToken)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", c.UserAgent)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != expectedReturnCode {
		return nil, errorFromJSONErrors(resp.Body)
	}

	return json.NewDecoder(resp.Body), nil
}

func (c *Client) genericJSONDataAPIPost(urlString string, dataPayload interface{}, authorizationToken string, expectedReturnCode int) error {
	return c.genericJSONDataAPIWriteNoResponse("POST", urlString, dataPayload, authorizationToken, expectedReturnCode)
}

func (c *Client) genericJSONDataAPIPut(urlString string, dataPayload interface{}, authorizationToken string, expectedReturnCode int) error {
	return c.genericJSONDataAPIWriteNoResponse("PUT", urlString, dataPayload, authorizationToken, expectedReturnCode)
}

func (c *Client) genericJSONDataAPIPostWithResponse(urlString string, dataPayload interface{}, authorizationToken string, expectedReturnCode int) (*json.Decoder, error) {
	return c.genericJSONDataAPIWriteWithResponse("POST", urlString, dataPayload, authorizationToken, expectedReturnCode)
}

func (c *Client) genericJSONDataAPIPutWithResponse(urlString string, dataPayload interface{}, authorizationToken string, expectedReturnCode int) (*json.Decoder, error) {
	return c.genericJSONDataAPIWriteWithResponse("PUT", urlString, dataPayload, authorizationToken, expectedReturnCode)
}

func (c *Client) genericJSONDataAPIWriteNoResponse(httpVerb string, urlString string, dataPayload interface{},
	authorizationToken string, expectedReturnCode int) error {
	decoder, err := c.genericJSONDataAPIWrite(httpVerb, urlString, dataPayload, authorizationToken, expectedReturnCode)
	if err != nil {
		return err
	}

	// When calling this function, we're discarding the response, but there might indeed have been
	// something in the body. To avoid screwing up our client, we need ensure the response
	// is drained and the body reader is closed.
	io.Copy(ioutil.Discard, decoder.Buffered())

	return nil
}

func (c *Client) genericJSONDataAPIWriteWithResponse(httpVerb string, urlString string, dataPayload interface{},
	authorizationToken string, expectedReturnCode int) (*json.Decoder, error) {
	decoder, err := c.genericJSONDataAPIWrite(httpVerb, urlString, dataPayload, authorizationToken, expectedReturnCode)
	if err != nil {
		return nil, err
	}

	return decoder, err
}

func (c *Client) genericJSONDataAPIWrite(httpVerb string, urlString string, dataPayload interface{},
	authorizationToken string, expectedReturnCode int) (*json.Decoder, error) {
	var requestBody struct {
		Data interface{} `json:"data"`
	}
	requestBody.Data = dataPayload

	b := new(bytes.Buffer)
	err := json.NewEncoder(b).Encode(requestBody)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(httpVerb, urlString, b)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", "Bearer "+authorizationToken)
	req.Header.Add("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", c.UserAgent)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != expectedReturnCode {
		return nil, errorFromJSONErrors(resp.Body)
	}

	return json.NewDecoder(resp.Body), nil
}

func (c *Client) genericJSONDataAPIDelete(urlString string, authorizationToken string, expectedReturnCode int) error {
	req, err := http.NewRequest("DELETE", urlString, nil)
	if err != nil {
		return err
	}
	req.Header.Add("Authorization", "Bearer "+authorizationToken)
	req.Header.Set("User-Agent", c.UserAgent)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode != expectedReturnCode {
		return errorFromJSONErrors(resp.Body)
	}

	// When calling this function, we're discarding the response, but there might indeed have been
	// something in the body. To avoid screwing up our client, we need ensure the response
	// is drained and the body reader is closed.
	io.Copy(ioutil.Discard, resp.Body)
	defer resp.Body.Close()

	return nil
}
