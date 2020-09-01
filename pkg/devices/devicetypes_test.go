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
	"connection-check/pkg/configuration"
	"connection-check/pkg/model"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
)

func TestCachedGetDeviceType(t *testing.T) {
	config, err := configuration.Load("../../config.json")
	if err != nil {
		t.Error(err)
		return
	}

	mux := sync.Mutex{}
	calls := []string{}

	mock := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mux.Lock()
		defer mux.Unlock()
		calls = append(calls, r.URL.Path)
		if strings.HasSuffix(r.URL.Path, "dt1") {
			json.NewEncoder(w).Encode(model.DeviceType{Id: "dt1", Name: "dt1name"})
		} else {
			http.Error(w, "not found", 404)
		}
	}))

	defer mock.Close()
	config.DeviceManagerUrl = mock.URL

	iot := New(config)

	t.Run("unknown device-type read 1", testGetDeviceTypeExpectError(iot, "unknown1"))
	t.Run("first device-type read", testGetDeviceType(iot, "dt1", "dt1name"))
	t.Run("second device-type read", testGetDeviceType(iot, "dt1", "dt1name"))
	t.Run("unknown device-type read 2", testGetDeviceTypeExpectError(iot, "unknown2"))
	t.Run("third device-type read", testGetDeviceType(iot, "dt1", "dt1name"))

	t.Run("check calls", func(t *testing.T) {
		mux.Lock()
		defer mux.Unlock()
		if len(calls) != 3 {
			t.Error(calls)
			return
		}

		if calls[0] != "/device-types/unknown1" {
			t.Error(calls[0])
			return
		}
		if calls[1] != "/device-types/dt1" {
			t.Error(calls[1])
			return
		}
		if calls[2] != "/device-types/unknown2" {
			t.Error(calls[2])
			return
		}
	})
}

func testGetDeviceType(iot *Devices, id string, expectedName string) func(t *testing.T) {
	return func(t *testing.T) {
		dt, err := iot.GetDeviceType("", id)
		if err != nil {
			t.Error(err)
			return
		}
		if dt.Id != id {
			t.Error(dt)
			return
		}
		if dt.Name != expectedName {
			t.Error(dt)
			return
		}
	}
}

func testGetDeviceTypeExpectError(iot *Devices, id string) func(t *testing.T) {
	return func(t *testing.T) {
		dt, err := iot.GetDeviceType("", id)
		if err == nil {
			t.Error(id, dt, err)
			return
		}
	}
}
