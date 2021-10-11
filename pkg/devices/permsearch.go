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
	return result, nil
}

//returns hubs as known by the permissions search service
func (this *Devices) ListHubs(token string, limit int, offset int) (result []model.Hub, err error) {
	req, err := http.NewRequest("GET", this.config.PermSearchUrl+"/jwt/list/hubs/r/"+strconv.Itoa(limit)+"/"+strconv.Itoa(offset)+"/name/asc", nil)
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
