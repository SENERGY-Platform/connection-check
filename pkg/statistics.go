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

package connectioncheck

import (
	"encoding/json"
	"time"
)

type Statistics struct {
	CheckedDevices            int `json:"checked_devices,omitempty"`
	CheckedHubs               int `json:"checked_hubs,omitempty"`
	ConnectedDevices          int `json:"connected_devices,omitempty"`
	ConnectedHubs             int `json:"connected_hubs,omitempty"`
	UpdateConnectedDevices    int `json:"update_connected_devices,omitempty"`
	UpdateDisconnectedDevices int `json:"update_disconnected_devices,omitempty"`
	UpdateConnectedHubs       int `json:"update_connected_hubs,omitempty"`
	UpdateDisconnectedHubs    int `json:"update_disconnected_hubs,omitempty"`
	timeVerneRequestsDevices  time.Duration
	timeVerneRequestsHubs     time.Duration
}

type PrintStatistics struct {
	Statistics
	TimeVerneRequestsDevices string `json:"time_verne_requests_devices,omitempty"`
	TimeVerneRequestsHubs    string `json:"time_verne_requests_hubs,omitempty"`
}

func (this *Statistics) AddCheckedDevices(count int) {
	if this != nil {
		this.CheckedDevices += count
	}
}

func (this *Statistics) AddCheckedHubs(count int) {
	if this != nil {
		this.CheckedHubs += count
	}
}

func (this *Statistics) AddConnectedHubs(count int) {
	if this != nil {
		this.ConnectedHubs += count
	}
}

func (this *Statistics) AddConnectedDevices(count int) {
	if this != nil {
		this.ConnectedDevices += count
	}
}

func (this *Statistics) AddUpdateConnectedDevices(count int) {
	if this != nil {
		this.UpdateConnectedDevices += count
	}
}

func (this *Statistics) AddUpdateDisconnectedDevices(count int) {
	if this != nil {
		this.UpdateDisconnectedDevices += count
	}
}

func (this *Statistics) AddUpdateConnectedHubs(count int) {
	if this != nil {
		this.UpdateConnectedHubs += count
	}
}

func (this *Statistics) AddUpdateDisconnectedHubs(count int) {
	if this != nil {
		this.UpdateDisconnectedHubs += count
	}
}

func (this *Statistics) AddTimeVerneRequestsDevices(dur time.Duration) {
	if this != nil {
		this.timeVerneRequestsDevices += dur
	}
}

func (this *Statistics) AddTimeVerneRequestsHubs(dur time.Duration) {
	if this != nil {
		this.timeVerneRequestsHubs += dur
	}
}

func (this *Statistics) String() string {
	if this != nil {
		timeDevices := ""
		if this.timeVerneRequestsDevices != 0 {
			timeDevices = this.timeVerneRequestsDevices.String()
		}
		timeHubs := ""
		if this.timeVerneRequestsHubs != 0 {
			timeHubs = this.timeVerneRequestsHubs.String()
		}
		temp, _ := json.Marshal(PrintStatistics{
			Statistics:               *this,
			TimeVerneRequestsDevices: timeDevices,
			TimeVerneRequestsHubs:    timeHubs,
		})
		return string(temp)
	}
	return ""
}
