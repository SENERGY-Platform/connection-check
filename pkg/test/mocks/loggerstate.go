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

import "sync"

func State() *LoggerStateMock {
	return &LoggerStateMock{
		HubStates:    map[string]bool{},
		DeviceStates: map[string]bool{},
		Mux:          &sync.Mutex{},
	}
}

type LoggerStateMock struct {
	Mux          *sync.Mutex
	HubStates    map[string]bool
	DeviceStates map[string]bool
}

func (this *LoggerStateMock) GetHubLogStates(token string, hubIds []string) (result map[string]bool, err error) {
	this.Mux.Lock()
	defer this.Mux.Unlock()
	result = map[string]bool{}
	for _, id := range hubIds {
		result[id] = this.HubStates[id]
	}
	return result, nil
}

func (this *LoggerStateMock) GetDeviceLogStates(token string, deviceIds []string) (result map[string]bool, err error) {
	this.Mux.Lock()
	defer this.Mux.Unlock()
	result = map[string]bool{}
	for _, id := range deviceIds {
		result[id] = this.DeviceStates[id]
	}
	return result, nil
}
