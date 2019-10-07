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
	"time"
)

// ResultSetOrder represents the order of the samples.
type ResultSetOrder int

const (
	// AscendingOrder means the Paginator will return results starting from the oldest.
	AscendingOrder ResultSetOrder = iota
	// DescendingOrder means the Paginator will return results starting from the oldest.
	DescendingOrder
)

// DatastreamPaginator handles a paginated set of results. It provides a one-directional iterator to call onto
// Astarte AppEngine API and handle potentially extremely large sets of results in chunk. You should prefer
// DatastreamPaginator rather than direct API calls if you expect your result set to be particularly large.
type DatastreamPaginator struct {
	baseURL        *url.URL
	windowStart    time.Time
	windowEnd      time.Time
	nextWindow     time.Time
	pageSize       int
	client         *Client
	token          string
	hasNextPage    bool
	resultSetOrder ResultSetOrder
}

// Rewind rewinds the simulator to the first page. GetNextPage will then return the first page of the call.
func (d *DatastreamPaginator) Rewind() {
	d.nextWindow = invalidTime
	d.hasNextPage = true
}

// HasNextPage returns whether this paginator can return more pages
func (d *DatastreamPaginator) HasNextPage() bool {
	return d.hasNextPage
}

// GetPageSize returns the page size for this paginator
func (d *DatastreamPaginator) GetPageSize() int {
	return d.pageSize
}

// GetResultSetOrder returns the order in which samples are returned for this paginator
func (d *DatastreamPaginator) GetResultSetOrder() ResultSetOrder {
	return d.resultSetOrder
}

// GetNextPage retrieves the next result page from the paginator. Returns the page as an array of DatastreamValue.
// If no more results are available, HasNextPage will return false. GetNextPage throws an error if no more pages are available.
func (d *DatastreamPaginator) GetNextPage() ([]DatastreamValue, error) {
	if !d.hasNextPage {
		return nil, errors.New("No more pages available")
	}

	callURL, _ := url.Parse(d.baseURL.String())
	queryString := ""
	if d.resultSetOrder == AscendingOrder {
		queryString += fmt.Sprintf("page_size=%v&to=%v", d.pageSize, d.windowEnd.UTC().Format(time.RFC3339Nano))
		if d.windowStart != invalidTime && d.nextWindow == invalidTime {
			queryString += fmt.Sprintf("&since=%v", d.windowStart.UTC().Format(time.RFC3339Nano))
		} else if d.nextWindow != invalidTime {
			queryString += fmt.Sprintf("&since_after=%v", d.nextWindow.UTC().Format(time.RFC3339Nano))
		}
	} else {
		queryString += fmt.Sprintf("limit=%v", d.pageSize)
		if d.windowStart != invalidTime {
			queryString += fmt.Sprintf("&since=%v", d.windowStart.UTC().Format(time.RFC3339Nano))
		}
		if d.nextWindow == invalidTime {
			queryString += fmt.Sprintf("&to=%v", d.windowEnd.UTC().Format(time.RFC3339Nano))
		} else {
			queryString += fmt.Sprintf("&to=%v", d.nextWindow.UTC().Format(time.RFC3339Nano))
		}
	}
	callURL.RawQuery = queryString

	decoder, err := d.client.genericJSONDataAPIGET(callURL.String(), d.token, 200)
	if err != nil {
		return nil, err
	}
	var responseBody struct {
		Data []DatastreamValue `json:"data"`
	}
	err = decoder.Decode(&responseBody)
	if err != nil {
		return nil, err
	}

	if len(responseBody.Data) < d.pageSize {
		d.hasNextPage = false
	} else {
		d.hasNextPage = true
		d.nextWindow = responseBody.Data[len(responseBody.Data)-1].Timestamp
	}

	return responseBody.Data, nil
}
