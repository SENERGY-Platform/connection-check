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
	"log"
	"net/http"
	"net/url"
	"runtime/debug"
)

func (this *Devices) GetDeviceByLocalId(token string, localId string) (result model.Device, err error) {
	err = this.cache.Use("local-devices."+localId, func() (interface{}, error) {
		return this.getDeviceByLocalId(token, localId)
	}, &result)
	return
}

func (this *Devices) cacheDevices(devices []model.Device) {
	for _, device := range devices {
		temp, err := json.Marshal(device)
		if err != nil {
			log.Println("ERROR:", err)
			debug.PrintStack()
			return
		}
		this.cache.Set("local-devices."+device.LocalId, temp)
	}
}

func (this *Devices) getDeviceByLocalId(token string, localId string) (result model.Device, err error) {
	req, err := http.NewRequest("GET", this.config.DeviceManagerUrl+"/local-devices/"+url.PathEscape(localId), nil)
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

func (this *Devices) HubContainsAnyGivenDeviceType(token string, cacheId string, hub model.Hub, dtIds []string) (result bool, err error) {
	err = this.cache.Use(cacheId, func() (interface{}, error) {
		return this.hubContainsAnyGivenDeviceType(token, hub, dtIds)
	}, &result)
	return
}

func (this *Devices) hubContainsAnyGivenDeviceType(token string, hub model.Hub, dtIds []string) (result bool, err error) {
	if dtIds == nil {
		return false, nil
	}
	if hub.DeviceLocalIds == nil {
		return false, nil
	}

	temp := []model.Device{}
	err = this.Query(token, QueryMessage{
		Resource: "devices",
		Find: &QueryFind{
			QueryListCommons: QueryListCommons{
				Limit:    1,
				Offset:   0,
				Rights:   "r",
				SortBy:   "name",
				SortDesc: false,
			},
			Filter: &Selection{
				And: []Selection{
					{
						Condition: ConditionConfig{
							Feature:   "features.device_type_id",
							Operation: QueryAnyValueInFeatureOperation,
							Value:     dtIds,
						},
					},
					{
						Condition: ConditionConfig{
							Feature:   "features.local_id",
							Operation: QueryAnyValueInFeatureOperation,
							Value:     hub.DeviceLocalIds,
						},
					},
				},
			},
		},
	}, &temp)
	if err != nil {
		return false, err
	}
	result = len(temp) > 0
	return result, nil
}
