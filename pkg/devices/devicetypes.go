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
)

func (this *Devices) GetDeviceType(token string, id string) (result model.DeviceType, err error) {
	err = this.cache.Use("device-type."+id, func() (interface{}, error) {
		return this.getDeviceType(token, id)
	}, &result)
	return
}

func (this *Devices) getDeviceType(token string, id string) (result model.DeviceType, err error) {
	req, err := http.NewRequest("GET", this.config.DeviceManagerUrl+"/device-types/"+url.PathEscape(id), nil)
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
		return result, errors.New(buf.String())
	}
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		debug.PrintStack()
		return result, err
	}
	return result, nil
}
