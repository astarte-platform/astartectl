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
	"time"

	"github.com/astarte-platform/astartectl/utils"
)

const defaultPageSize int = 10000

var invalidTime time.Time = time.Unix(0, 0)

// AppEngineService is the API Client for AppEngine API
type AppEngineService struct {
	client       *Client
	appEngineURL *url.URL
}

// DeviceIdentifierType represents what kind of identifier is used for identifying a Device.
type DeviceIdentifierType int

const (
	// AutodiscoverDeviceIdentifier is the default, and uses heuristics to autodetermine which kind of
	// identifier is being used.
	AutodiscoverDeviceIdentifier DeviceIdentifierType = iota
	// AstarteDeviceID is the Device's ID in its standard format.
	AstarteDeviceID
	// AstarteDeviceAlias is one of the Device's Aliases.
	AstarteDeviceAlias
)

func resolveDevicePath(deviceIdentifier string, deviceIdentifierType DeviceIdentifierType) string {
	switch deviceIdentifierType {
	case AutodiscoverDeviceIdentifier:
		if utils.IsValidAstarteDeviceID(deviceIdentifier) {
			return fmt.Sprintf("devices/%v", deviceIdentifier)
		}
		return fmt.Sprintf("devices-by-alias/%v", deviceIdentifier)
	case AstarteDeviceID:
		return fmt.Sprintf("devices/%v", deviceIdentifier)
	case AstarteDeviceAlias:
		return fmt.Sprintf("devices-by-alias/%v", deviceIdentifier)
	}
	return ""
}

// ListDevices returns a list of Devices in the Realm
func (s *AppEngineService) ListDevices(realm string, token string) ([]string, error) {
	callURL, _ := url.Parse(s.appEngineURL.String())
	callURL.Path = path.Join(callURL.Path, fmt.Sprintf("/v1/%s/devices", realm))
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

// GetDevice returns the DeviceDetails of a single Device in the Realm
func (s *AppEngineService) GetDevice(realm string, deviceIdentifier string, deviceIdentifierType DeviceIdentifierType, token string) (DeviceDetails, error) {
	callURL, _ := url.Parse(s.appEngineURL.String())
	callURL.Path = path.Join(callURL.Path, fmt.Sprintf("/v1/%s/%s", realm, resolveDevicePath(deviceIdentifier, deviceIdentifierType)))
	decoder, err := s.client.genericJSONDataAPIGET(callURL.String(), token, 200)
	if err != nil {
		return DeviceDetails{}, err
	}
	var responseBody struct {
		Data DeviceDetails `json:"data"`
	}
	err = decoder.Decode(&responseBody)
	if err != nil {
		return DeviceDetails{}, err
	}

	return responseBody.Data, nil
}

// ListDeviceInterfaces returns the list of Interfaces exposed by the Device's introspection
func (s *AppEngineService) ListDeviceInterfaces(realm string, deviceIdentifier string,
	deviceIdentifierType DeviceIdentifierType, token string) ([]string, error) {
	callURL, _ := url.Parse(s.appEngineURL.String())
	callURL.Path = path.Join(callURL.Path, fmt.Sprintf("/v1/%s/%s/interfaces", realm, resolveDevicePath(deviceIdentifier, deviceIdentifierType)))
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

// ListDeviceAliases is an helper to list all aliases of a Device
func (s *AppEngineService) ListDeviceAliases(realm string, deviceID string, token string) (map[string]string, error) {
	deviceDetails, err := s.GetDevice(realm, deviceID, AstarteDeviceID, token)
	if err != nil {
		return nil, err
	}
	return deviceDetails.Aliases, nil
}

func parsePropertyInterface(interfaceMap map[string]interface{}) map[string]interface{} {
	// Start recursion and return resulting map
	return parsePropertiesMap(interfaceMap, "")
}

func parseDatastreamInterface(interfaceMap map[string]interface{}) (map[string]DatastreamValue, error) {
	// Start recursion and return resulting map
	return parseDatastreamMap(interfaceMap, "")
}

func parseAggregateDatastreamInterface(interfaceMap map[string]interface{}) (DatastreamAggregateValue, error) {
	// Start recursion and return resulting map
	return DatastreamAggregateValue{}, nil
}

func parsePropertiesMap(aMap map[string]interface{}, completeKeyPath string) map[string]interface{} {
	m := make(map[string]interface{})

	for key, val := range aMap {
		switch actualVal := val.(type) {
		case map[string]interface{}:
			for k, v := range parsePropertiesMap(val.(map[string]interface{}), completeKeyPath+"/"+key) {
				m[k] = v
			}
		default:
			m[completeKeyPath+"/"+key] = actualVal
		}
	}

	return m
}

func parseDatastreamMap(aMap map[string]interface{}, completeKeyPath string) (map[string]DatastreamValue, error) {
	m := make(map[string]DatastreamValue)

	// Special case: have we hit the bottom?
	if _, ok := aMap["value"]; ok {
		datastreamValue, err := parseDatastreamValue(aMap)
		if err != nil {
			return nil, err
		}
		m[completeKeyPath] = datastreamValue
		return m, nil
	}

	for key, val := range aMap {
		switch val.(type) {
		case map[string]interface{}:
			parsedMap, err := parseDatastreamMap(val.(map[string]interface{}), completeKeyPath+"/"+key)
			if err != nil {
				return nil, err
			}
			for k, v := range parsedMap {
				m[k] = v
			}
		}
	}

	return m, nil
}

func parseDatastreamValue(aMap map[string]interface{}) (DatastreamValue, error) {
	// Ensure some type safety
	switch aMap["timestamp"].(type) {
	case time.Time:
		return DatastreamValue{Value: aMap["value"], Timestamp: aMap["timestamp"].(time.Time),
			ReceptionTimestamp: aMap["reception_timestamp"].(time.Time)}, nil
	case string:
		timestamp, err := time.Parse(time.RFC3339Nano, aMap["timestamp"].(string))
		if err != nil {
			return DatastreamValue{}, err
		}
		receptionTimestamp, _ := time.Parse(time.RFC3339Nano, aMap["reception_timestamp"].(string))
		if err != nil {
			return DatastreamValue{}, err
		}
		return DatastreamValue{Value: aMap["value"], Timestamp: timestamp, ReceptionTimestamp: receptionTimestamp}, nil
	}

	return DatastreamValue{}, errors.New("Unable to parse Datastream")
}

// GetProperties returns all the currently set Properties on a given Interface
func (s *AppEngineService) GetProperties(realm string, deviceIdentifier string, deviceIdentifierType DeviceIdentifierType,
	interfaceName string, token string) (map[string]interface{}, error) {
	callURL, _ := url.Parse(s.appEngineURL.String())
	callURL.Path = path.Join(callURL.Path, fmt.Sprintf("/v1/%s/%s/interfaces/%s", realm,
		resolveDevicePath(deviceIdentifier, deviceIdentifierType), interfaceName))
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

	return parsePropertyInterface(responseBody.Data), nil
}

// GetDatastreamSnapshot returns all the last values on all paths for a Datastream interface
func (s *AppEngineService) GetDatastreamSnapshot(realm string, deviceIdentifier string, deviceIdentifierType DeviceIdentifierType,
	interfaceName string, token string) (map[string]DatastreamValue, error) {
	callURL, _ := url.Parse(s.appEngineURL.String())
	callURL.Path = path.Join(callURL.Path, fmt.Sprintf("/v1/%s/%s/interfaces/%s", realm,
		resolveDevicePath(deviceIdentifier, deviceIdentifierType), interfaceName))
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

	return parseDatastreamInterface(responseBody.Data)
}

// GetLastDatastreams returns all the last values on a path for a Datastream interface.
// If limit is <= 0, it returns all existing datastreams. Consider using a GetDatastreamsPaginator in that case.
func (s *AppEngineService) GetLastDatastreams(realm string, deviceIdentifier string, deviceIdentifierType DeviceIdentifierType, interfaceName string, interfacePath string, limit int, token string) ([]DatastreamValue, error) {
	return s.getDatastreamInternal(realm, resolveDevicePath(deviceIdentifier, deviceIdentifierType), interfaceName, interfacePath, invalidTime, invalidTime, limit, DescendingOrder, token)
}

// GetDatastreamsPaginator returns a Paginator for all the values on a path for a Datastream interface.
func (s *AppEngineService) GetDatastreamsPaginator(realm string, deviceIdentifier string, deviceIdentifierType DeviceIdentifierType, interfaceName string, interfacePath string, resultSetOrder ResultSetOrder, token string) DatastreamPaginator {
	return s.getDatastreamPaginatorInternal(realm, resolveDevicePath(deviceIdentifier, deviceIdentifierType), interfaceName, interfacePath, invalidTime, time.Now(), defaultPageSize, resultSetOrder, token)
}

// GetDatastreamsTimeWindowPaginator returns a Paginator for all the values on a path in a specified time window for a Datastream interface.
func (s *AppEngineService) GetDatastreamsTimeWindowPaginator(realm string, deviceIdentifier string, deviceIdentifierType DeviceIdentifierType, interfaceName string, interfacePath string, since time.Time, to time.Time, resultSetOrder ResultSetOrder, token string) DatastreamPaginator {
	return s.getDatastreamPaginatorInternal(realm, resolveDevicePath(deviceIdentifier, deviceIdentifierType), interfaceName, interfacePath, since, to, defaultPageSize, resultSetOrder, token)
}

// GetAggregateDatastreamSnapshot returns the last value for a Datastream aggregate interface
func (s *AppEngineService) GetAggregateDatastreamSnapshot(realm string, deviceIdentifier string, deviceIdentifierType DeviceIdentifierType, interfaceName string, token string) (DatastreamAggregateValue, error) {
	callURL, _ := url.Parse(s.appEngineURL.String())
	callURL.Path = path.Join(callURL.Path, fmt.Sprintf("/v1/%s/%s/interfaces/%s", realm, resolveDevicePath(deviceIdentifier, deviceIdentifierType), interfaceName))
	// It's a snapshot, so limit=1
	callURL.RawQuery = "limit=1"
	decoder, err := s.client.genericJSONDataAPIGET(callURL.String(), token, 200)
	if err != nil {
		return DatastreamAggregateValue{}, err
	}
	var responseBody struct {
		Data []map[string]interface{} `json:"data"`
	}
	err = decoder.Decode(&responseBody)
	if err != nil {
		return DatastreamAggregateValue{}, err
	}

	// If there is no data, return an empty value
	if len(responseBody.Data) == 0 {
		return DatastreamAggregateValue{}, nil
	}

	return parseAggregateDatastreamInterface(responseBody.Data[0])
}

// GetLastAggregateDatastreams returns the last count values for a Datastream aggregate interface
func (s *AppEngineService) GetLastAggregateDatastreams(realm string, deviceIdentifier string, deviceIdentifierType DeviceIdentifierType, interfaceName string, token string, count int) ([]DatastreamAggregateValue, error) {
	callURL, _ := url.Parse(s.appEngineURL.String())
	callURL.Path = path.Join(callURL.Path, fmt.Sprintf("/v1/%s/%s/interfaces/%s", realm, resolveDevicePath(deviceIdentifier, deviceIdentifierType), interfaceName))
	callURL.RawQuery = fmt.Sprintf("limit=%v", count)
	decoder, err := s.client.genericJSONDataAPIGET(callURL.String(), token, 200)
	if err != nil {
		return nil, err
	}
	var responseBody struct {
		Data []DatastreamAggregateValue `json:"data"`
	}
	err = decoder.Decode(&responseBody)
	if err != nil {
		return nil, err
	}

	return responseBody.Data, nil
}

// GetAggregateDatastreamsTimeWindow returns the last count values for a Datastream aggregate interface
func (s *AppEngineService) GetAggregateDatastreamsTimeWindow(realm string, deviceIdentifier string, deviceIdentifierType DeviceIdentifierType, interfaceName string, token string, since time.Time, to time.Time) ([]DatastreamAggregateValue, error) {
	callURL, _ := url.Parse(s.appEngineURL.String())
	callURL.Path = path.Join(callURL.Path, fmt.Sprintf("/v1/%s/%s/interfaces/%s", realm, resolveDevicePath(deviceIdentifier, deviceIdentifierType), interfaceName))
	// It's a snapshot, so limit=1
	callURL.RawQuery = fmt.Sprintf("since=%s&to=%s", since.UTC().Format(time.RFC3339Nano), to.UTC().Format(time.RFC3339Nano))
	decoder, err := s.client.genericJSONDataAPIGET(callURL.String(), token, 200)
	if err != nil {
		return nil, err
	}
	var responseBody struct {
		Data []DatastreamAggregateValue `json:"data"`
	}
	err = decoder.Decode(&responseBody)
	if err != nil {
		return nil, err
	}

	return responseBody.Data, nil
}

// AddDeviceAlias adds an Alias to a Device
func (s *AppEngineService) AddDeviceAlias(realm string, deviceID string, aliasTag string, deviceAlias string, token string) error {
	callURL, _ := url.Parse(s.appEngineURL.String())
	callURL.Path = path.Join(callURL.Path, fmt.Sprintf("/v1/%s/devices/%s", realm, deviceID))
	payload := map[string]map[string]string{"aliases": {aliasTag: deviceAlias}}
	err := s.client.genericJSONDataAPIPatch(callURL.String(), payload, token, 200)
	if err != nil {
		return err
	}

	return nil
}

// DeleteDeviceAlias deletes an Alias from a Device based on the Alias' tag
func (s *AppEngineService) DeleteDeviceAlias(realm string, deviceID string, aliasTag string, token string) error {
	callURL, _ := url.Parse(s.appEngineURL.String())
	callURL.Path = path.Join(callURL.Path, fmt.Sprintf("/v1/%s/devices/%s", realm, deviceID))
	// We're using map[string]interface{} rather than map[string]string since we want to have null
	// rather than an empty string in the JSON payload, and this is the only way.
	payload := map[string]map[string]interface{}{"aliases": {aliasTag: nil}}
	err := s.client.genericJSONDataAPIPatch(callURL.String(), payload, token, 200)
	if err != nil {
		return err
	}

	return nil
}

func (s *AppEngineService) getDatastreamInternal(realm string, devicePath string, interfaceName string, interfacePath string,
	since time.Time, to time.Time, limit int, resultSetOrder ResultSetOrder, token string) ([]DatastreamValue, error) {
	realLimit := limit
	if limit < 0 || limit > defaultPageSize {
		realLimit = defaultPageSize
	}
	datastreamPaginator := s.getDatastreamPaginatorInternal(realm, devicePath, interfaceName, interfacePath, since, to, realLimit, resultSetOrder, token)

	var resultSet []DatastreamValue
	for ok := true; ok; ok = datastreamPaginator.HasNextPage() {
		page, err := datastreamPaginator.GetNextPage()
		if err != nil {
			return nil, err
		}

		// Check special cases
		if limit > 0 {
			totalSize := len(resultSet) + len(page)
			if totalSize == limit {
				return append(resultSet, page...), nil
			} else if totalSize > limit {
				missingSamples := limit - len(resultSet)
				return append(resultSet, page[0:missingSamples-1]...), nil
			}
		}

		resultSet = append(resultSet, page...)
	}

	return resultSet, nil
}

func (s *AppEngineService) getDatastreamPaginatorInternal(realm string, devicePath string, interfaceName string, interfacePath string,
	since time.Time, to time.Time, pageSize int, resultSetOrder ResultSetOrder, token string) DatastreamPaginator {
	callURL, _ := url.Parse(s.appEngineURL.String())
	callURL.Path = path.Join(callURL.Path, fmt.Sprintf("/v1/%s/%s/interfaces/%s%s", realm, devicePath, interfaceName, interfacePath))

	datastreamPaginator := DatastreamPaginator{
		baseURL:        callURL,
		windowStart:    since,
		windowEnd:      to,
		nextWindow:     invalidTime,
		pageSize:       pageSize,
		client:         s.client,
		token:          token,
		hasNextPage:    true,
		resultSetOrder: resultSetOrder,
	}
	return datastreamPaginator
}
