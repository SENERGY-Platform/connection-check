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
	security "connection-check/pkg/auth"
	"connection-check/pkg/configuration"
	"connection-check/pkg/connectionlog/logger"
	"connection-check/pkg/connectionlog/state"
	"connection-check/pkg/devices"
	"connection-check/pkg/model"
	"connection-check/pkg/topicgenerator"
	"connection-check/pkg/topicgenerator/common"
	"connection-check/pkg/vernemq"
	"context"
	"errors"
	"log"
	"strings"
	"time"
)

func New(config configuration.Config) (*ConnectionCheck, error) {
	topic, ok := topicgenerator.Known[config.TopicGenerator]
	if !ok {
		return nil, errors.New("unknown topic generator " + config.TopicGenerator)
	}
	logger, err := logger.New(config.ZookeeperUrl, true, false, config.DeviceLogTopic, config.HubLogTopic)
	if err != nil {
		return nil, err
	}
	handledProtocols := map[string]bool{}
	for _, protocolId := range config.HandledProtocols {
		handledProtocols[strings.TrimSpace(protocolId)] = true
	}
	return &ConnectionCheck{
		Logger:                     logger,
		LoggerState:                state.New(config.ConnectionLogStateUrl),
		Verne:                      vernemq.New(config.VernemqManagementUrl),
		Devices:                    devices.New(config),
		TokenGen:                   security.New(config.AuthEndpoint, config.AuthClientId, config.AuthClientSecret, 2),
		SubscriptionTopicGenerator: topic,
		BatchSize:                  config.BatchSize,
		HandledProtocols:           handledProtocols,
		Debug:                      config.Debug,
	}, nil
}

type ConnectionCheck struct {
	Logger                     Logger
	LoggerState                LoggerState
	Verne                      Verne
	Devices                    Devices
	TokenGen                   TokenGenerator
	SubscriptionTopicGenerator TopicGenerator
	BatchSize                  int
	HandledProtocols           map[string]bool
	Debug                      bool
}

func (this *ConnectionCheck) RunInterval(ctx context.Context, duration time.Duration) {
	go func() {
		ticker := time.NewTicker(duration)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				this.runDevices()
				this.runHubs()
			}
		}
	}()
}

func (this *ConnectionCheck) runDevices() {
	startTime := time.Now()
	log.Println("start device-check")
	err := this.RunDevices()
	log.Println("finish device-check", time.Now().Sub(startTime), err)
}

func (this *ConnectionCheck) runHubs() {
	startTime := time.Now()
	log.Println("start hub-check")
	err := this.RunHubs()
	log.Println("finish hub-check", time.Now().Sub(startTime), err)
}

func (this *ConnectionCheck) RunDevices() (err error) {
	limit := this.BatchSize
	offset := 0
	count := limit
	for count == limit {
		count, err = this.RunDeviceBatch(limit, offset)
		if err != nil {
			return err
		}
		offset = offset + limit
	}
	return nil
}

func (this *ConnectionCheck) RunHubs() (err error) {
	limit := this.BatchSize
	offset := 0
	count := limit
	for count == limit {
		count, err = this.RunHubBatch(limit, offset)
		if err != nil {
			return err
		}
		offset = offset + limit
	}
	return nil
}

func (this *ConnectionCheck) RunHubBatch(limit int, offset int) (count int, err error) {
	token, err := this.TokenGen.Access()
	if err != nil {
		return count, err
	}
	hubs, err := this.Devices.ListHubs(token, limit, offset)
	if err != nil {
		return count, err
	}
	ids := []string{}
	filteredHubs := []model.Hub{}
	for _, hub := range hubs {
		if this.hubMatchesHandledProtocols(token, hub) {
			ids = append(ids, hub.Id)
			filteredHubs = append(filteredHubs, hub)
		}
	}
	onlineStates, err := this.LoggerState.GetHubLogStates(token, ids)
	if err != nil {
		return count, err
	}
	for _, hub := range filteredHubs {
		subscriptionIsOnline, err := this.Verne.CheckOnlineClient(hub.Id)
		if err != nil {
			return count, err
		}
		deviceHasOnlineState := onlineStates[hub.Id]

		if deviceHasOnlineState && !subscriptionIsOnline {
			err = this.Logger.LogHubDisconnect(hub.Id)
		}
		if !deviceHasOnlineState && subscriptionIsOnline {
			err = this.Logger.LogHubConnect(hub.Id)
		}
		if err != nil {
			return count, err
		}
	}
	return len(hubs), nil
}

func (this *ConnectionCheck) RunDeviceBatch(limit int, offset int) (count int, err error) {
	token, err := this.TokenGen.Access()
	if err != nil {
		return count, err
	}
	devices, err := this.Devices.ListDevices(token, limit, offset)
	if err != nil {
		return count, err
	}
	ids := []string{}
	for _, device := range devices {
		ids = append(ids, device.Id)
	}
	onlineStates, err := this.LoggerState.GetDeviceLogStates(token, ids)
	if err != nil {
		return count, err
	}
	dtCache := map[string]model.DeviceType{}
	for _, device := range devices {
		dt, ok := dtCache[device.DeviceTypeId]
		if !ok {
			dt, err = this.Devices.GetDeviceType(token, device.DeviceTypeId)
			if err != nil {
				return count, err
			}
			dtCache[dt.Id] = dt
		}
		topic, err := this.SubscriptionTopicGenerator(device, dt, this.HandledProtocols)
		if err == common.NoSubscriptionExpected {
			err = nil
			continue
		}
		if err != nil {
			return count, err
		}
		subscriptionIsOnline, err := this.Verne.CheckOnlineSubscription(topic)
		if err != nil {
			return count, err
		}
		deviceHasOnlineState := onlineStates[device.Id]

		if deviceHasOnlineState && !subscriptionIsOnline {
			err = this.Logger.LogDeviceDisconnect(device.Id)
		}
		if !deviceHasOnlineState && subscriptionIsOnline {
			err = this.Logger.LogDeviceConnect(device.Id)
		}
		if err != nil {
			return count, err
		}
	}
	return len(devices), nil
}

func (this *ConnectionCheck) hubMatchesHandledProtocols(token string, hub model.Hub) bool {
	dtCache := map[string]model.DeviceType{}
	for _, deviceLocalId := range hub.DeviceLocalIds {
		device, err := this.Devices.GetDeviceByLocalId(token, deviceLocalId)
		if err != nil && this.Debug {
			log.Println("WARNING: hubMatchesHandledProtocols() unable to load device", deviceLocalId, err)
		}
		if err == nil {
			dt, ok := dtCache[device.DeviceTypeId]
			if !ok {
				dt, err = this.Devices.GetDeviceType(token, device.DeviceTypeId)
				if err != nil {
					if this.Debug {
						log.Println("WARNING: hubMatchesHandledProtocols() unable to load device-type", device.DeviceTypeId, err)
					}
				} else {
					dtCache[dt.Id] = dt
				}
			}
			if err == nil {
				if common.DeviceTypeUsesHandledProtocol(dt, this.HandledProtocols) {
					return true
				}
			}
		}
	}
	return false
}

func (this *ConnectionCheck) deviceTypeMatchesHandledProtocols(dt model.DeviceType) bool {
	for _, service := range dt.Services {
		if this.HandledProtocols[service.ProtocolId] {
			return true
		}
	}
	return false
}
