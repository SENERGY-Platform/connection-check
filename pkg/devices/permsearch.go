/*
 * Copyright 2020 InfAI (CC SES)
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package devices

import (
	"bytes"
	"connection-check/pkg/model"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"runtime/debug"
	"strconv"
)

func (this *Devices) ListDevicesAfter(token string, limit int, after model.Device) (result []model.Device, err error) {
	req, err := http.NewRequest("GET", this.config.PermSearchUrl+"/v3/resources/devices?limit="+strconv.Itoa(limit)+"&after.id="+url.QueryEscape(after.Id)+"&after.sort_field_value="+url.QueryEscape(`"`+after.Name+`"`), nil)
	if err != nil {
		debug.PrintStack()
		return result, err
	}
	req.Header.Set("Authorization", token)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		debug.PrintStack()
		return result, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		buf := new(bytes.Buffer)
		buf.ReadFrom(resp.Body)
		debug.PrintStack()
		return result, errors.New(buf.String())
	}
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		debug.PrintStack()
		return result, err
	}
	this.cacheDevices(result)
	return result, nil
}

//returns devices as known by the permissions search service
func (this *Devices) ListDevices(token string, limit int, offset int) (result []model.Device, err error) {
	req, err := http.NewRequest("GET", this.config.PermSearchUrl+"/v3/resources/devices?limit="+strconv.Itoa(limit)+"&offset="+strconv.Itoa(offset)+"&sort=name&rights=r", nil)
	if err != nil {
		debug.PrintStack()
		return result, err
	}
	req.Header.Set("Authorization", token)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		debug.PrintStack()
		return result, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		buf := new(bytes.Buffer)
		buf.ReadFrom(resp.Body)
		debug.PrintStack()
		return result, errors.New(buf.String())
	}
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		debug.PrintStack()
		return result, err
	}
	this.cacheDevices(result)
	return result, nil
}

//returns hubs as known by the permissions search service
func (this *Devices) ListHubs(token string, limit int, offset int) (result []model.Hub, err error) {
	req, err := http.NewRequest("GET", this.config.PermSearchUrl+"/v3/resources/hubs?limit="+strconv.Itoa(limit)+"&offset="+strconv.Itoa(offset)+"&sort=name&rights=r", nil)
	if err != nil {
		debug.PrintStack()
		return result, err
	}
	req.Header.Set("Authorization", token)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		debug.PrintStack()
		return result, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		buf := new(bytes.Buffer)
		buf.ReadFrom(resp.Body)
		debug.PrintStack()
		return result, errors.New(buf.String())
	}
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		debug.PrintStack()
		return result, err
	}
	return result, nil
}

func (this *Devices) Query(token string, query QueryMessage, result interface{}) (err error) {
	body := new(bytes.Buffer)
	err = json.NewEncoder(body).Encode(query)
	if err != nil {
		return err
	}
	req, err := http.NewRequest("POST", this.config.PermSearchUrl+"/v3/query", body)
	if err != nil {
		debug.PrintStack()
		return err
	}
	req.Header.Set("Authorization", token)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		debug.PrintStack()
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		buf := new(bytes.Buffer)
		buf.ReadFrom(resp.Body)
		debug.PrintStack()
		return errors.New(buf.String())
	}
	err = json.NewDecoder(resp.Body).Decode(result)
	if err != nil {
		debug.PrintStack()
		return err
	}
	return nil
}

type QueryMessage struct {
	Resource      string     `json:"resource"`
	Find          *QueryFind `json:"find"`
	TermAggregate *string    `json:"term_aggregate"`
}
type QueryFind struct {
	QueryListCommons
	Search string     `json:"search"`
	Filter *Selection `json:"filter"`
}

type QueryListCommons struct {
	Limit    int        `json:"limit"`
	Offset   int        `json:"offset"`
	After    *ListAfter `json:"after"`
	Rights   string     `json:"rights"`
	SortBy   string     `json:"sort_by"`
	SortDesc bool       `json:"sort_desc"`
}

type ListAfter struct {
	SortFieldValue interface{} `json:"sort_field_value"`
	Id             string      `json:"id"`
}

type QueryOperationType string

const (
	QueryEqualOperation             QueryOperationType = "=="
	QueryUnequalOperation           QueryOperationType = "!="
	QueryAnyValueInFeatureOperation QueryOperationType = "any_value_in_feature"
)

type ConditionConfig struct {
	Feature   string             `json:"feature"`
	Operation QueryOperationType `json:"operation"`
	Value     interface{}        `json:"value"`
	Ref       string             `json:"ref"`
}

type Selection struct {
	And       []Selection     `json:"and"`
	Or        []Selection     `json:"or"`
	Not       *Selection      `json:"not"`
	Condition ConditionConfig `json:"condition"`
}
