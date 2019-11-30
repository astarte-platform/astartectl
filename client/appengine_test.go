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
	"encoding/json"
	"reflect"
	"testing"
)

// Test data
const uglyLookingDatastream = "{\"data\":{\"nested\":{\"value\":{\"value\":15,\"timestamp\":\"2019-01-01T01:23:45.678Z\",\"reception_timestamp\":\"2019-01-01T01:23:45.678Z\"}, \"timestamp\":{\"value\":\"something\",\"timestamp\":\"2019-01-01T01:23:45.678Z\",\"reception_timestamp\":\"2019-01-01T01:23:45.678Z\"}}}}"
const uglyLookingProperties = "{\"data\":{\"nested\":{\"value\":15, \"timestamp\":\"not_a_timestamp\"}}}"

func TestParametricDatastreamParsing(t *testing.T) {
	// Get the parametric datastream and treat it as individual
	var responseBody struct {
		Data map[string]interface{} `json:"data"`
	}
	err := json.Unmarshal([]byte(uglyLookingDatastream), &responseBody)
	if err != nil {
		t.Error(err)
	}

	val, err := parseDatastreamInterface(responseBody.Data)
	if err != nil {
		t.Error(err)
	}

	if ds, ok := val["/nested/value"]; ok {
		if v, ok := ds.Value.(float64); !ok || v != 15 {
			t.Error("Error in parsing /nested/value", ok, ds, v, reflect.TypeOf(ds))
		}
	} else {
		t.Error("Error in parsing /nested/value", val)
	}

	if ds, ok := val["/nested/timestamp"]; ok {
		if v, ok := ds.Value.(string); !ok || v != "something" {
			t.Error("Error in parsing /nested/timestamp", ok, ds, v, reflect.TypeOf(ds))
		}
	} else {
		t.Error("Error in parsing /nested/timestamp", val)
	}
}

func TestPropertiesParsing(t *testing.T) {
	// Get the parametric datastream and treat it as individual
	var responseBody struct {
		Data map[string]interface{} `json:"data"`
	}
	err := json.Unmarshal([]byte(uglyLookingProperties), &responseBody)
	if err != nil {
		t.Error(err)
	}

	val := parsePropertyInterface(responseBody.Data)

	if ds, ok := val["/nested/value"]; ok {
		if v, ok := ds.(float64); !ok || v != 15 {
			t.Error("Error in parsing /nested/value", ok, ds, v, reflect.TypeOf(ds))
		}
	} else {
		t.Error("Error in parsing /nested/value", val)
	}

	if ds, ok := val["/nested/timestamp"]; ok {
		if v, ok := ds.(string); !ok || v != "not_a_timestamp" {
			t.Error("Error in parsing /nested/timestamp", ok, ds, v, reflect.TypeOf(ds))
		}
	} else {
		t.Error("Error in parsing /nested/timestamp", val)
	}
}
