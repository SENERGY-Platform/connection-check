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
	Checked                int `json:"checked"`
	Connected              int `json:"connected"`
	UpdateConnected        int `json:"update_connected"`
	UpdateDisconnected     int `json:"update_disconnected"`
	timeVerneRequests      time.Duration
	timeListRequests       time.Duration
	timeRequestDeviceTypes time.Duration
	timeRequestLocalDevice time.Duration
	timeRequestLogState    time.Duration
}

type PrintStatistics struct {
	Statistics
	TimeVerneRequests      string `json:"time_verne_requests,omitempty"`
	TimeListRequests       string `json:"time_list_requests,omitempty"`
	TimeRequestDeviceTypes string `json:"time_request_device_types,omitempty"`
	TimeRequestLocalDevice string `json:"time_request_local_device,omitempty"`
	TimeRequestLogState    string `json:"time_request_log_state,omitempty"`
}

func (this *Statistics) AddChecked(count int) {
	if this != nil {
		this.Checked += count
	}
}

func (this *Statistics) AddConnected(count int) {
	if this != nil {
		this.Connected += count
	}
}

func (this *Statistics) AddUpdateConnected(count int) {
	if this != nil {
		this.UpdateConnected += count
	}
}

func (this *Statistics) AddUpdateDisconnected(count int) {
	if this != nil {
		this.UpdateDisconnected += count
	}
}

func (this *Statistics) AddTimeVerneRequests(dur time.Duration) {
	if this != nil {
		this.timeVerneRequests += dur
	}
}

func (this *Statistics) AddTimeListRequests(dur time.Duration) {
	if this != nil {
		this.timeListRequests += dur
	}
}

func (this *Statistics) AddTimeRequestDeviceTypes(dur time.Duration) {
	if this != nil {
		this.timeRequestDeviceTypes += dur
	}
}

func (this *Statistics) AddTimeRequestLocalDevice(dur time.Duration) {
	if this != nil {
		this.timeRequestLocalDevice += dur
	}
}

func (this *Statistics) AddTimeRequestLogState(dur time.Duration) {
	if this != nil {
		this.timeRequestLogState += dur
	}
}

func (this *Statistics) String() string {
	if this != nil {
		timeVerneRequests := ""
		if this.timeVerneRequests != 0 {
			timeVerneRequests = this.timeVerneRequests.String()
		}
		timeListRequests := ""
		if this.timeListRequests != 0 {
			timeListRequests = this.timeListRequests.String()
		}
		timeRequestDeviceTypes := ""
		if this.timeRequestDeviceTypes != 0 {
			timeRequestDeviceTypes = this.timeRequestDeviceTypes.String()
		}
		timeRequestLocalDevice := ""
		if this.timeRequestLocalDevice != 0 {
			timeRequestLocalDevice = this.timeRequestLocalDevice.String()
		}
		timeRequestLogState := ""
		if this.timeRequestLocalDevice != 0 {
			timeRequestLogState = this.timeRequestLogState.String()
		}
		temp, _ := json.Marshal(PrintStatistics{
			Statistics:             *this,
			TimeVerneRequests:      timeVerneRequests,
			TimeListRequests:       timeListRequests,
			TimeRequestDeviceTypes: timeRequestDeviceTypes,
			TimeRequestLocalDevice: timeRequestLocalDevice,
			TimeRequestLogState:    timeRequestLogState,
		})
		return string(temp)
	}
	return ""
}
