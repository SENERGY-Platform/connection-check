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

import "connection-check/pkg/model"

type Logger interface {
	LogDeviceDisconnect(deviceId string) error
	LogDeviceConnect(deviceId string) error
	LogHubConnect(clientId string) error
	LogHubDisconnect(clientId string) error
}

type LoggerState interface {
	GetHubLogStates(token string, hubIds []string) (result map[string]bool, err error)
	GetDeviceLogStates(token string, deviceIds []string) (result map[string]bool, err error)
}

type TokenGenerator interface {
	Access() (token string, err error)
}

type Devices interface {
	GetDeviceType(token string, id string) (result model.DeviceType, err error)
	ListHubs(token string, limit int, offset int) (result []model.Hub, err error)
	ListDevices(token string, limit int, offset int) (result []model.Device, err error)
	ListDevicesAfter(token string, limit int, after model.Device) (result []model.Device, err error)
	ListAllDeviceTypesWithFilter(token string, cacheId string, filter func(dt model.DeviceType) bool) ([]model.DeviceType, error)
	HubContainsAnyGivenDeviceType(token string, cacheId string, hub model.Hub, dtIds []string) (bool, error)
}

type Verne interface {
	CheckOnlineSubscription(topic string) (onlineSubscriptionExists bool, err error)
	CheckOnlineSubscriptions(topics []string) (onlineSubscriptionExists bool, err error)
	CheckOnlineClient(clientId string) (onlineClientExists bool, err error)
}

type TopicGenerator = func(device model.Device, deviceType model.DeviceType, handledProtocols map[string]bool) (topicCandidates []string, err error)
