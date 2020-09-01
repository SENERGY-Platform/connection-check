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

package common

import (
	"connection-check/pkg/model"
	"strings"
)

func GetHandledServices(services []model.Service, handledProtocols map[string]bool) (result []model.Service) {
	for _, service := range services {
		if handledProtocols[service.ProtocolId] && UsesControllingFunction(service) {
			result = append(result, service)
		}
	}
	return
}

func DeviceTypeUsesHandledProtocol(dt model.DeviceType, handledProtocols map[string]bool) (result bool) {
	for _, service := range dt.Services {
		if handledProtocols[service.ProtocolId] {
			return true
		}
	}
	return false
}

func UsesControllingFunction(service model.Service) bool {
	for _, function := range service.FunctionIds {
		if IsControllingFunctionId(function) {
			return true
		}
	}
	return false
}

func IsControllingFunctionId(id string) bool {
	if strings.HasPrefix(id, model.CONTROLLING_FUNCTION_PREFIX) {
		return true
	}
	return false
}
