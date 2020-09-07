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

package vernemq

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"runtime/debug"
	"strconv"
)

func New(url string) *VernemqManagementApi {
	return &VernemqManagementApi{
		Url:             url,
		NodeResultLimit: 100,
	}
}

type VernemqManagementApi struct {
	Url             string
	NodeResultLimit int
}

func (this *VernemqManagementApi) GetOnlineClients() (result []Client, err error) {
	path := "/api/v1/session/show?--is_online=true&--client_id&--user&--limit=" + strconv.Itoa(this.NodeResultLimit)
	req, err := http.NewRequest("GET", this.Url+path, nil)
	if err != nil {
		debug.PrintStack()
		return result, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		debug.PrintStack()
		return result, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		buf, _ := ioutil.ReadAll(resp.Body)
		err = errors.New(resp.Status + ":" + string(buf))
		log.Println("ERROR: unable to get resource", err)
		return result, err
	}
	temp := ClientWrapper{}
	err = json.NewDecoder(resp.Body).Decode(&temp)
	if err != nil {
		log.Println("ERROR: unable to unmarshal result of", this.Url+path)
		return result, err
	}
	return temp.Table, nil
}

func (this *VernemqManagementApi) GetOnlineSubscriptions() (result []Subscription, err error) {
	path := "/api/v1/session/show?--is_online=true&--user&--client_id&--topic&--limit=" + strconv.Itoa(this.NodeResultLimit)
	req, err := http.NewRequest("GET", this.Url+path, nil)
	if err != nil {
		debug.PrintStack()
		return result, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		debug.PrintStack()
		return result, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		buf, _ := ioutil.ReadAll(resp.Body)
		err = errors.New(resp.Status + ":" + string(buf))
		log.Println("ERROR: unable to get result from vernemq", err)
		return result, err
	}
	temp := SubscriptionWrapper{}
	err = json.NewDecoder(resp.Body).Decode(&temp)
	if err != nil {
		log.Println("ERROR: unable to unmarshal result of", this.Url+path)
		return result, err
	}
	return temp.Table, nil
}

func (this *VernemqManagementApi) CheckOnlineSubscriptions(topics []string) (onlineSubscriptionExists bool, err error) {
	for _, topic := range topics {
		onlineSubscriptionExists, err = this.CheckOnlineSubscription(topic)
		if err != nil {
			return
		}
		if onlineSubscriptionExists {
			return
		}
	}
	return
}

func (this *VernemqManagementApi) CheckOnlineSubscription(topic string) (onlineSubscriptionExists bool, err error) {
	path := "/api/v1/session/show?--is_online=true&--topic=" + url.QueryEscape(topic) + "&--limit=1"
	req, err := http.NewRequest("GET", this.Url+path, nil)
	if err != nil {
		debug.PrintStack()
		return false, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		debug.PrintStack()
		return false, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		buf, _ := ioutil.ReadAll(resp.Body)
		err = errors.New(resp.Status + ":" + string(buf))
		log.Println("ERROR: unable to get result from vernemq", err)
		return false, err
	}
	temp := SubscriptionWrapper{}
	err = json.NewDecoder(resp.Body).Decode(&temp)
	if err != nil {
		log.Println("ERROR: unable to unmarshal result of", this.Url+path)
		return false, err
	}
	return len(temp.Table) > 0, nil
}

func (this *VernemqManagementApi) CheckOnlineClient(clientId string) (onlineClientExists bool, err error) {
	path := "/api/v1/session/show?--is_online=true&--client_id=" + url.QueryEscape(clientId) + "&--limit=1"
	req, err := http.NewRequest("GET", this.Url+path, nil)
	if err != nil {
		debug.PrintStack()
		return false, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		debug.PrintStack()
		return false, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		buf, _ := ioutil.ReadAll(resp.Body)
		err = errors.New(resp.Status + ":" + string(buf))
		log.Println("ERROR: unable to get result from vernemq", err)
		return false, err
	}
	temp := SubscriptionWrapper{}
	err = json.NewDecoder(resp.Body).Decode(&temp)
	if err != nil {
		log.Println("ERROR: unable to unmarshal result of", this.Url+path)
		return false, err
	}
	return len(temp.Table) > 0, nil
}
