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
	"connection-check/pkg/model"
	"connection-check/pkg/test/docker"
	"connection-check/pkg/test/mocks"
	"connection-check/pkg/topicgenerator"
	"connection-check/pkg/vernemq"
	"context"
	"encoding/json"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/ory/dockertest/v3"
	"reflect"
	"sync"
	"testing"
	"time"
)

func TestConnectionCheck(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	wg := &sync.WaitGroup{}
	defer wg.Wait()
	defer cancel()

	pool, err := dockertest.NewPool("")
	if err != nil {
		t.Error(err)
		return
	}

	brokerUrl, managementUrl, err := docker.VernemqWithManagementApi(pool, ctx, wg)
	if err != nil {
		t.Error(err)
		return
	}

	loggerMock := mocks.Logger()
	stateMock := mocks.State()
	iotMock := mocks.Devices()

	check := ConnectionCheck{
		Logger:                     loggerMock,
		LoggerState:                stateMock,
		Verne:                      vernemq.New(managementUrl),
		Devices:                    iotMock,
		TokenGen:                   mocks.TokenGen,
		SubscriptionTopicGenerator: topicgenerator.Known["senergy"],
		BatchSize:                  2,
		HandledProtocols:           map[string]bool{"test-protocol": true},
		Debug:                      true,
	}
	t.Run("create device type 1", testCreateDeviceType(iotMock, "dt1", []model.Service{
		{
			LocalId:     "sl2",
			ProtocolId:  "nope",
			FunctionIds: []string{model.CONTROLLING_FUNCTION_PREFIX + "f1"},
		},
		{
			LocalId:     "sl1",
			ProtocolId:  "test-protocol",
			FunctionIds: []string{model.CONTROLLING_FUNCTION_PREFIX + "f1"},
		},
		{
			LocalId:     "sl3",
			ProtocolId:  "test-protocol",
			FunctionIds: []string{model.MEASURING_FUNCTION_PREFIX + "nope"},
		},
	}))

	t.Run("create device type nope 1", testCreateDeviceType(iotMock, "dt_nope1", []model.Service{
		{
			LocalId:     "sl2",
			ProtocolId:  "nope",
			FunctionIds: []string{model.CONTROLLING_FUNCTION_PREFIX + "f1"},
		},
	}))

	t.Run("create device type nope 2", testCreateDeviceType(iotMock, "dt_nope2", []model.Service{
		{
			LocalId:     "sl3",
			ProtocolId:  "test-protocol",
			FunctionIds: []string{model.MEASURING_FUNCTION_PREFIX + "nope"},
		},
	}))

	t.Run("create true online iot device", testCreateDevice(iotMock, stateMock, "true_online", "dt1", true))
	t.Run("create true offline iot device", testCreateDevice(iotMock, stateMock, "true_offline", "dt1", false))
	t.Run("create false online iot device", testCreateDevice(iotMock, stateMock, "false_online", "dt1", true))
	t.Run("create false offline iot device", testCreateDevice(iotMock, stateMock, "false_offline", "dt1", false))
	t.Run("create false offline iot device", testCreateDevice(iotMock, stateMock, "false_offline_2", "dt1", false))

	t.Run("create ignored device 1", testCreateDevice(iotMock, stateMock, "ignore1", "dt_nope1", true))
	t.Run("create ignored device 2", testCreateDevice(iotMock, stateMock, "ignore2", "dt_nope2", true))
	t.Run("create ignored device 3", testCreateDevice(iotMock, stateMock, "ignore3", "dt_nope1", false))
	t.Run("create ignored device 4", testCreateDevice(iotMock, stateMock, "ignore4", "dt_nope2", false))

	t.Run("create true online hub", testCreateHub(iotMock, stateMock, "true_online_hub", []string{"true_online", "true_offline"}, true))
	t.Run("create false online hub", testCreateHub(iotMock, stateMock, "false_online_hub", []string{"true_online", "true_offline"}, true))
	t.Run("create true offline hub", testCreateHub(iotMock, stateMock, "true_offline_hub", []string{"true_online", "true_offline"}, false))
	t.Run("create false offline hub", testCreateHub(iotMock, stateMock, "false_offline_hub", []string{"true_online", "true_offline"}, false))

	t.Run("create ignored hub 1", testCreateHub(iotMock, stateMock, "ignored_hub_1", []string{"ignore1", "ignore3"}, true))
	t.Run("create ignored hub 2", testCreateHub(iotMock, stateMock, "ignored_hub_2", []string{}, true))
	t.Run("create ignored hub 3", testCreateHub(iotMock, stateMock, "ignored_hub_3", []string{"ignore1", "ignore3"}, false))
	t.Run("create ignored hub 4", testCreateHub(iotMock, stateMock, "ignored_hub_4", []string{}, false))

	t.Run("create not actually ignored hub", testCreateHub(iotMock, stateMock, "not_actually_ignored", []string{"ignore1", "ignore2"}, true))

	t.Run("create mqtt client true_online_hub", testStartTestClient(ctx, wg, brokerUrl, "true_online_hub", []string{
		"command/true_online/sl1",
		"command/false_offline/sl1",
	}))

	t.Run("create mqtt client false_offline_hub", testStartTestClient(ctx, wg, brokerUrl, "false_offline_hub", []string{"command/false_offline_2/sl1"}))

	time.Sleep(2 * time.Second)

	t.Run("run devices", func(t *testing.T) {
		err = check.RunDevices(nil)
		if err != nil {
			t.Error(err)
			return
		}
	})

	t.Run("run hubs", func(t *testing.T) {
		err = check.RunHubs(nil)
		if err != nil {
			t.Error(err)
			return
		}
	})

	t.Run("check events", func(t *testing.T) {
		expected := []mocks.LogEvent{
			{Id: "false_online", Kind: "device", Connected: false},
			{Id: "false_offline", Kind: "device", Connected: true},
			{Id: "false_offline_2", Kind: "device", Connected: true},
			{Id: "false_online_hub", Kind: "hub", Connected: false},
			{Id: "false_offline_hub", Kind: "hub", Connected: true},
			{Id: "not_actually_ignored", Kind: "hub", Connected: false},
		}

		if !reflect.DeepEqual(loggerMock.Events, expected) {
			temp, _ := json.Marshal(loggerMock.Events)
			t.Error(string(temp))
		}
	})

	t.Run("run private debug", func(t *testing.T) {
		check.Debug = true
		check.runDevices(nil)
		check.runHubs(nil)
	})
	t.Run("run private without debug", func(t *testing.T) {
		check.Debug = false
		check.runDevices(nil)
		check.runHubs(nil)
	})
}

func TestConnectionCheckPlaceholder(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	wg := &sync.WaitGroup{}
	defer wg.Wait()
	defer cancel()

	pool, err := dockertest.NewPool("")
	if err != nil {
		t.Error(err)
		return
	}

	brokerUrl, managementUrl, err := docker.VernemqWithManagementApi(pool, ctx, wg)
	if err != nil {
		t.Error(err)
		return
	}

	loggerMock := mocks.Logger()
	stateMock := mocks.State()
	iotMock := mocks.Devices()

	check := ConnectionCheck{
		Logger:                     loggerMock,
		LoggerState:                stateMock,
		Verne:                      vernemq.New(managementUrl),
		Devices:                    iotMock,
		TokenGen:                   mocks.TokenGen,
		SubscriptionTopicGenerator: topicgenerator.Known["senergy"],
		BatchSize:                  2,
		HandledProtocols:           map[string]bool{"test-protocol": true},
		Debug:                      true,
	}
	t.Run("create device with +", testCreateDeviceType(iotMock, "dt1", []model.Service{
		{
			LocalId:     "sl1",
			ProtocolId:  "test-protocol",
			FunctionIds: []string{model.CONTROLLING_FUNCTION_PREFIX + "f1"},
		},
	}))

	t.Run("create true online device", testCreateDevice(iotMock, stateMock, "true_online", "dt1", true))
	t.Run("create true online + device", testCreateDevice(iotMock, stateMock, "true_online_p", "dt1", true))
	t.Run("create true online # device", testCreateDevice(iotMock, stateMock, "true_online_s", "dt1", true))

	t.Run("create true offline + device", testCreateDevice(iotMock, stateMock, "true_offline", "dt1", false))

	t.Run("create false online device", testCreateDevice(iotMock, stateMock, "false_online", "dt1", true))

	t.Run("create false offline device", testCreateDevice(iotMock, stateMock, "false_offline", "dt1", false))
	t.Run("create false offline + device", testCreateDevice(iotMock, stateMock, "false_offline_p", "dt1", false))
	t.Run("create false offline # device", testCreateDevice(iotMock, stateMock, "false_offline_s", "dt1", false))

	t.Run("create mqtt client", testStartTestClient(ctx, wg, brokerUrl, "hub", []string{
		"command/true_online/sl1",
		"command/true_online_p/+",
		"command/true_online_s/#",
		"command/false_offline/sl1",
		"command/false_offline_p/+",
		"command/false_offline_s/#",
	}))

	time.Sleep(2 * time.Second)

	t.Run("run devices", func(t *testing.T) {
		err = check.RunDevices(nil)
		if err != nil {
			t.Error(err)
			return
		}
	})

	t.Run("check events", func(t *testing.T) {
		expected := []mocks.LogEvent{
			{Id: "false_online", Kind: "device", Connected: false},
			{Id: "false_offline", Kind: "device", Connected: true},
			{Id: "false_offline_p", Kind: "device", Connected: true},
			{Id: "false_offline_s", Kind: "device", Connected: true},
		}

		if !reflect.DeepEqual(loggerMock.Events, expected) {
			temp, _ := json.Marshal(loggerMock.Events)
			t.Error(string(temp))
		}
	})
}

func testCreateHub(iot *mocks.DevicesMock, stateRepo *mocks.LoggerStateMock, id string, devices []string, state bool) func(t *testing.T) {
	return func(t *testing.T) {
		iot.Mux.Lock()
		defer iot.Mux.Unlock()
		stateRepo.Mux.Lock()
		defer stateRepo.Mux.Unlock()
		iot.Hubs = append(iot.Hubs, model.Hub{
			Id:             id,
			DeviceLocalIds: devices,
		})
		stateRepo.HubStates[id] = state
	}
}

func testCreateDevice(iot *mocks.DevicesMock, stateRepo *mocks.LoggerStateMock, localId string, deviceType string, state bool) func(t *testing.T) {
	return func(t *testing.T) {
		iot.Mux.Lock()
		defer iot.Mux.Unlock()
		stateRepo.Mux.Lock()
		defer stateRepo.Mux.Unlock()
		id := localId
		iot.Devices = append(iot.Devices, model.Device{
			Id:           id,
			LocalId:      localId,
			DeviceTypeId: deviceType,
		})
		stateRepo.DeviceStates[id] = state
	}
}

func testCreateDeviceType(iot *mocks.DevicesMock, id string, services []model.Service) func(t *testing.T) {
	return func(t *testing.T) {
		iot.Mux.Lock()
		defer iot.Mux.Unlock()
		iot.DeviceTypes = append(iot.DeviceTypes, model.DeviceType{
			Id:       id,
			Services: services,
		})
	}
}

func testStartTestClient(ctx context.Context, wg *sync.WaitGroup, broker string, clientId string, subscriptions []string) func(t *testing.T) {
	return func(t *testing.T) {
		options := mqtt.NewClientOptions().
			SetCleanSession(true).
			SetClientID(clientId).
			SetUsername("test").
			SetPassword("test").
			SetAutoReconnect(true).
			AddBroker(broker)

		client := mqtt.NewClient(options)
		if token := client.Connect(); token.Wait() && token.Error() != nil {
			t.Error(token.Error())
			return
		}
		wg.Add(1)
		go func() {
			<-ctx.Done()
			client.Disconnect(0)
			wg.Done()
		}()
		for _, topic := range subscriptions {
			token := client.Subscribe(topic, 0, func(client mqtt.Client, message mqtt.Message) {})
			if token.Wait() && token.Error() != nil {
				t.Error(token.Error())
				return
			}
		}
	}
}
