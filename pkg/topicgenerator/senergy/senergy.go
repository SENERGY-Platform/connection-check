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

package senergy

import (
	"connection-check/pkg/model"
	"connection-check/pkg/topicgenerator/common"
	"connection-check/pkg/topicgenerator/known"
)

func init() {
	known.Generators["senergy"] = func(device model.Device, deviceType model.DeviceType, handledProtocols map[string]bool) (topicCandidates []string, err error) {
		services := common.GetHandledServices(deviceType.Services, handledProtocols)
		if len(services) == 0 {
			return topicCandidates, common.NoSubscriptionExpected
		}
		for _, service := range services {
			topicCandidates = append(topicCandidates, "command/"+device.LocalId+"/"+service.LocalId)
		}
		topicCandidates = append(topicCandidates, "command/"+device.LocalId+"/+")
		topicCandidates = append(topicCandidates, "command/"+device.LocalId+"/#")
		return topicCandidates, nil
	}
}
