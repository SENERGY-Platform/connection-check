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

package mocks

import (
	"connection-check/pkg/model"
	"errors"
	"sync"
)

func Devices() *DevicesMock {
	return &DevicesMock{
		DeviceTypes: []model.DeviceType{},
		Devices:     []model.Device{},
		Hubs:        []model.Hub{},
		Mux:         &sync.Mutex{},
	}
}

type DevicesMock struct {
	Mux         *sync.Mutex
	DeviceTypes []model.DeviceType
	Devices     []model.Device
	Hubs        []model.Hub
}

func (this *DevicesMock) ListDevicesAfter(token string, limit int, after model.Device) (result []model.Device, err error) {
	this.Mux.Lock()
	defer this.Mux.Unlock()
	afterFoundAt := -1
	for i, d := range this.Devices {
		if d.Id == after.Id {
			afterFoundAt = i
			break
		}
	}
	if afterFoundAt > -1 {
		offset := afterFoundAt + 1
		end := offset + limit
		if end > len(this.Devices) {
			end = len(this.Devices)
		}
		result = this.Devices[offset:end]
	}
	return
}

func (this *DevicesMock) GetDeviceType(token string, id string) (result model.DeviceType, err error) {
	this.Mux.Lock()
	defer this.Mux.Unlock()
	for _, dt := range this.DeviceTypes {
		if dt.Id == id {
			return dt, nil
		}
	}
	return result, errors.New("device-type not found '" + id + "'")
}

func (this *DevicesMock) GetDeviceByLocalId(token string, localId string) (result model.Device, err error) {
	this.Mux.Lock()
	defer this.Mux.Unlock()
	for _, device := range this.Devices {
		if device.LocalId == localId {
			return device, nil
		}
	}
	return result, errors.New("device not found")
}

func (this *DevicesMock) ListHubs(token string, limit int, offset int) (result []model.Hub, err error) {
	this.Mux.Lock()
	defer this.Mux.Unlock()
	if offset >= len(this.Hubs) {
		return result, nil
	}
	end := offset + limit
	if end > len(this.Hubs) {
		end = len(this.Hubs)
	}
	return this.Hubs[offset:end], nil
}

func (this *DevicesMock) ListDevices(token string, limit int, offset int) (result []model.Device, err error) {
	this.Mux.Lock()
	defer this.Mux.Unlock()
	if offset >= len(this.Devices) {
		return result, nil
	}
	end := offset + limit
	if end > len(this.Devices) {
		end = len(this.Devices)
	}
	return this.Devices[offset:end], nil
}
