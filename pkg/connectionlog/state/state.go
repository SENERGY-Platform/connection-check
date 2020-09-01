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

package state

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"runtime/debug"
)

func New(url string) *ConnectionLogState {
	return &ConnectionLogState{url: url}
}

type ConnectionLogState struct {
	url string
}

func (this *ConnectionLogState) GetDeviceLogStates(token string, deviceIds []string) (result map[string]bool, err error) {
	b := new(bytes.Buffer)
	err = json.NewEncoder(b).Encode(deviceIds)
	if err != nil {
		return result, err
	}
	req, err := http.NewRequest("POST", this.url+"/intern/state/device/check", b)
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

func (this *ConnectionLogState) GetHubLogStates(token string, hubIds []string) (result map[string]bool, err error) {
	b := new(bytes.Buffer)
	err = json.NewEncoder(b).Encode(hubIds)
	if err != nil {
		return result, err
	}
	req, err := http.NewRequest("POST", this.url+"/intern/state/gateway/check", b)
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
