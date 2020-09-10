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
	"sync"
	"time"
)

func NewHealthChecker(expectedCheckInterval time.Duration, errorLimit int) *HealthChecker {
	return &HealthChecker{
		expectedCheckInterval: expectedCheckInterval,
		hubErrCount:           0,
		lastIntervalStart:     time.Now(),
		errorLimit:            errorLimit,
	}
}

type HealthChecker struct {
	expectedCheckInterval time.Duration
	lastIntervalStart     time.Time
	hubErrCount           int
	deviceErrCount        int
	mux                   sync.Mutex
	errorLimit            int
}

func (this *HealthChecker) Check() (ok bool, info interface{}) {
	if this == nil {
		return false, "health checker not initialized (this == nil)"
	}
	this.mux.Lock()
	defer this.mux.Unlock()
	info = map[string]interface{}{"hubErrCount": this.hubErrCount, "deviceErrCount": this.deviceErrCount, "lastIntervalStart": this.lastIntervalStart}
	age := time.Since(this.lastIntervalStart)
	if this.hubErrCount > this.errorLimit || this.deviceErrCount > this.errorLimit || (this.expectedCheckInterval > 0 && age > this.expectedCheckInterval) {
		return false, info
	} else {
		return true, info
	}
}

func (this *HealthChecker) LogIntervalStart() {
	if this == nil {
		return
	}
	this.mux.Lock()
	defer this.mux.Unlock()
	this.lastIntervalStart = time.Now()
}

func (this *HealthChecker) LogErrorHubs(err error) {
	if this == nil {
		return
	}
	if err == nil {
		this.hubErrCount = 0
	} else {
		this.hubErrCount = this.hubErrCount + 1
	}
}

func (this *HealthChecker) LogErrorDevices(err error) {
	if this == nil {
		return
	}
	if err == nil {
		this.deviceErrCount = 0
	} else {
		this.deviceErrCount = this.deviceErrCount + 1
	}
}
