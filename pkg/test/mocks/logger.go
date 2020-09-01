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

func Logger() *LoggerMock {
	return &LoggerMock{
		Mux:    &sync.Mutex{},
		Events: []LogEvent{},
	}
}

type LogEvent struct {
	Id        string
	Kind      string
	Connected bool
}

type LoggerMock struct {
	Mux    *sync.Mutex
	Events []LogEvent
}

func (this *LoggerMock) LogDeviceDisconnect(deviceId string) error {
	this.Mux.Lock()
	defer this.Mux.Unlock()
	this.Events = append(this.Events, LogEvent{
		Id:        deviceId,
		Kind:      "device",
		Connected: false,
	})
	return nil
}

func (this *LoggerMock) LogDeviceConnect(deviceId string) error {
	this.Mux.Lock()
	defer this.Mux.Unlock()
	this.Events = append(this.Events, LogEvent{
		Id:        deviceId,
		Kind:      "device",
		Connected: true,
	})
	return nil
}

func (this *LoggerMock) LogHubConnect(clientId string) error {
	this.Mux.Lock()
	defer this.Mux.Unlock()
	this.Events = append(this.Events, LogEvent{
		Id:        clientId,
		Kind:      "hub",
		Connected: true,
	})
	return nil
}

func (this *LoggerMock) LogHubDisconnect(clientId string) error {
	this.Mux.Lock()
	defer this.Mux.Unlock()
	this.Events = append(this.Events, LogEvent{
		Id:        clientId,
		Kind:      "hub",
		Connected: false,
	})
	return nil
}
