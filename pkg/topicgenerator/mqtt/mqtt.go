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
	known.Generators["mqtt"] = func(device model.Device, deviceType model.DeviceType, handledProtocols map[string]bool) (topicCandidates []string, err error) {
		services := common.GetHandledServices(deviceType.Services, handledProtocols)
		if len(services) == 0 {
			return topicCandidates, common.NoSubscriptionExpected
		}

		gen := topic.New(DefaultActuatorPattern)

		oneLevelPlaceholderTopic, err := gen.Create(device.Id, "+")
		if err != nil {
			return topicCandidates, err
		}
		multiLevelPlaceholderTopic, err := gen.Create(device.Id, "#")
		if err != nil {
			return topicCandidates, err
		}

		set := map[string]bool{
			oneLevelPlaceholderTopic:   true,
			multiLevelPlaceholderTopic: true,
		}
		for _, service := range services {
			topic, err := gen.Create(device.Id, service.LocalId)
			if err != nil {
				return topicCandidates, err
			}
			set[topic] = true
		}
		for topic, _ := range set {
			topicCandidates = append(topicCandidates, topic)
		}
		return topicCandidates, nil
	}
}
