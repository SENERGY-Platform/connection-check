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
	"runtime/debug"
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
	assignmentIndex, err := ParseAssignmentId(config.AssignmentId)
	if err != nil {
		return nil, err
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
		AssignmentIndex:            assignmentIndex,
		Scaling:                    config.Scaling,
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
	intervalContext            context.Context
	AssignmentIndex            int
	Scaling                    int
}

func (this *ConnectionCheck) RunInterval(ctx context.Context, duration time.Duration, health *HealthChecker) {
	go func() {
		this.run(health)
		ticker := time.NewTicker(duration)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				this.run(health)
			}
		}
	}()
}

func (this *ConnectionCheck) run(health *HealthChecker) {
	health.LogIntervalStart()
	this.runDevices(health)
	this.runHubs(health)
}

func (this *ConnectionCheck) runDevices(health *HealthChecker) {
	startTime := time.Now()

	var statistics *Statistics
	if this.Debug {
		statistics = &Statistics{}
	}

	log.Println("start device-check")
	err := this.RunDevices(statistics)
	health.LogErrorDevices(err)
	log.Println("finish device-check", err, time.Since(startTime), statistics.String())
}

func (this *ConnectionCheck) runHubs(health *HealthChecker) {
	startTime := time.Now()

	var statistics *Statistics
	if this.Debug {
		statistics = &Statistics{}
	}

	log.Println("start hub-check")
	err := this.RunHubs(statistics)
	health.LogErrorHubs(err)
	log.Println("finish hub-check", err, time.Since(startTime), statistics.String())
}

func (this *ConnectionCheck) RunDevices(statistics *Statistics) (err error) {
	limit := this.BatchSize
	offset := 0
	count := limit
	var last *model.Device
	for count == limit {
		last, count, err = this.RunDeviceBatch(limit, offset, last, statistics)
		if err != nil {
			return err
		}
		offset = offset + limit
	}
	return nil
}

func (this *ConnectionCheck) RunHubs(statistics *Statistics) (err error) {
	limit := this.BatchSize
	offset := 0
	count := limit
	for count == limit {
		isAssignment := IsAssignedBatch(this.BatchSize, offset, this.Scaling, this.AssignmentIndex)
		if isAssignment {
			count, err = this.RunHubBatch(limit, offset, statistics)
			if err != nil {
				return err
			}
		}
		offset = offset + limit
	}
	return nil
}

func (this *ConnectionCheck) RunHubBatch(limit int, offset int, statistics *Statistics) (count int, err error) {
	if this.Debug {
		log.Println("DEBUG: run hub batch", limit, offset)
	}
	token, err := this.TokenGen.Access()
	if err != nil {
		return count, err
	}
	listStart := time.Now()
	hubs, err := this.Devices.ListHubs(token, limit, offset)
	if err != nil {
		return count, err
	}
	if this.Debug {
		log.Println("DEBUG: hubs", len(hubs))
	}
	statistics.AddTimeListRequests(time.Since(listStart))
	ids := []string{}
	filteredHubs := []model.Hub{}
	for _, hub := range hubs {
		if this.hubMatchesHandledProtocols(token, hub, statistics) {
			ids = append(ids, hub.Id)
			filteredHubs = append(filteredHubs, hub)
		}
	}
	if this.Debug {
		log.Println("DEBUG: filtered hubs", len(filteredHubs))
	}
	logStateStart := time.Now()
	onlineStates, err := this.LoggerState.GetHubLogStates(token, ids)
	if err != nil {
		return count, err
	}
	if this.Debug {
		log.Println("DEBUG: got states of filtered hubs")
	}
	statistics.AddTimeRequestLogState(time.Since(logStateStart))
	statistics.AddChecked(len(filteredHubs))
	for _, hub := range filteredHubs {
		if this.Debug {
			log.Println("DEBUG: check hub", hub.Id)
		}
		timeVerneStart := time.Now()
		subscriptionIsOnline, err := this.Verne.CheckOnlineClient(hub.Id)
		if err != nil {
			return count, err
		}
		statistics.AddTimeVerneRequests(time.Since(timeVerneStart))
		if subscriptionIsOnline {
			statistics.AddConnected(1)
		}

		hubHasOnlineState := onlineStates[hub.Id]

		if hubHasOnlineState && !subscriptionIsOnline {
			statistics.AddUpdateDisconnected(1)
			err = this.Logger.LogHubDisconnect(hub.Id)
			if this.Debug {
				log.Println("DEBUG: connect hub", hub)
			}
		}
		if !hubHasOnlineState && subscriptionIsOnline {
			statistics.AddUpdateConnected(1)
			err = this.Logger.LogHubConnect(hub.Id)
			if this.Debug {
				log.Println("DEBUG: disconnect hub", hub)
			}
		}
		if err != nil {
			return count, err
		}
	}
	return len(hubs), nil
}

func (this *ConnectionCheck) RunDeviceBatch(limit int, offset int, after *model.Device, statistics *Statistics) (last *model.Device, count int, err error) {
	if this.Debug {
		log.Println("DEBUG: run device batch", limit, offset, after)
	}
	token, err := this.TokenGen.Access()
	if err != nil {
		return last, count, err
	}

	listStart := time.Now()
	var devices []model.Device
	if after != nil {
		devices, err = this.Devices.ListDevicesAfter(token, limit, *after)
	} else {
		devices, err = this.Devices.ListDevices(token, limit, offset)
	}
	if err != nil {
		return last, count, err
	}
	if len(devices) > 0 {
		last = &devices[len(devices)-1]
	}
	isAssignment := IsAssignedBatch(this.BatchSize, offset, this.Scaling, this.AssignmentIndex)
	if isAssignment {
		statistics.AddTimeListRequests(time.Since(listStart))
		ids := []string{}
		for _, device := range devices {
			ids = append(ids, device.Id)
		}
		logStateStart := time.Now()
		onlineStates, err := this.LoggerState.GetDeviceLogStates(token, ids)
		if err != nil {
			return last, count, err
		}
		statistics.AddTimeRequestLogState(time.Since(logStateStart))
		dtCache := map[string]model.DeviceType{}
		for _, device := range devices {
			dt, ok := dtCache[device.DeviceTypeId]
			if !ok {
				dtStart := time.Now()
				dt, err = this.Devices.GetDeviceType(token, device.DeviceTypeId)
				if err != nil {
					return last, count, err
				}
				statistics.AddTimeRequestDeviceTypes(time.Since(dtStart))
				dtCache[dt.Id] = dt
			}
			topics, err := this.SubscriptionTopicGenerator(device, dt, this.HandledProtocols)
			if err == common.NoSubscriptionExpected {
				err = nil
				continue
			}
			if err != nil {
				return last, count, err
			}
			statistics.AddChecked(1)
			timeVerneStart := time.Now()
			subscriptionIsOnline, err := this.Verne.CheckOnlineSubscriptions(topics)
			if err != nil {
				return last, count, err
			}
			statistics.AddTimeVerneRequests(time.Since(timeVerneStart))

			if subscriptionIsOnline {
				statistics.AddConnected(1)
			}

			deviceHasOnlineState := onlineStates[device.Id]

			if deviceHasOnlineState && !subscriptionIsOnline {
				statistics.AddUpdateDisconnected(1)
				err = this.Logger.LogDeviceDisconnect(device.Id)
				if this.Debug {
					log.Println("DEBUG: disconnect device", device)
				}
			}
			if !deviceHasOnlineState && subscriptionIsOnline {
				statistics.AddUpdateConnected(1)
				err = this.Logger.LogDeviceConnect(device.Id)
				if this.Debug {
					log.Println("DEBUG: connect device", device)
				}
			}
			if err != nil {
				return last, count, err
			}
		}
	}
	return last, len(devices), nil
}

func (this *ConnectionCheck) hubMatchesHandledProtocols(token string, hub model.Hub, statistics *Statistics) bool {
	deviceTypes, err := this.Devices.ListAllDeviceTypesWithFilter(token, "dt.list.filtered", func(dt model.DeviceType) bool { return common.DeviceTypeUsesHandledProtocol(dt, this.HandledProtocols) })
	if err != nil {
		log.Println("ERROR:", err)
		debug.PrintStack()
		return false
	}
	dtIds := []string{}
	for _, dt := range deviceTypes {
		dtIds = append(dtIds, dt.Id)
	}
	result, err := this.Devices.HubContainsAnyGivenDeviceType(token, "hub.check.dts."+hub.Id, hub, dtIds)
	if err != nil {
		log.Println("ERROR:", err)
		debug.PrintStack()
		return false
	}
	return result
}
