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

package mqtt

import (
	"connection-check/pkg/model"
	"connection-check/pkg/topicgenerator/common"
	"connection-check/pkg/topicgenerator/known"
	"connection-check/pkg/topicgenerator/mqtt/topic"
)

const DefaultActuatorPattern = "{{.DeviceId}}/cmnd/{{.LocalServiceId}}"

func init() {
	known.Generators["mqtt"] = func(device model.Device, deviceType model.DeviceType, handledProtocols map[string]bool) (result string, err error) {
		services := common.GetHandledServices(deviceType.Services, handledProtocols)
		if len(services) == 0 {
			return result, common.NoSubscriptionExpected
		}
		service := services[0]
		return topic.New(DefaultActuatorPattern).Create(device.Id, service.LocalId)
	}
}
