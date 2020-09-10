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
	"connection-check/pkg/health"
	"connection-check/pkg/model"
	"connection-check/pkg/test/docker"
	"connection-check/pkg/test/mocks"
	"connection-check/pkg/topicgenerator"
	"connection-check/pkg/vernemq"
	"context"
	"github.com/ory/dockertest/v3"
	"io/ioutil"
	"net/http"
	"strconv"
	"sync"
	"testing"
	"time"
)

func TestHealthCheck(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	wg := &sync.WaitGroup{}
	defer wg.Wait()
	defer cancel()

	pool, err := dockertest.NewPool("")
	if err != nil {
		t.Error(err)
		return
	}

	_, managementUrl, err := docker.VernemqWithManagementApi(pool, ctx, wg)
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

	healthChecker := NewHealthChecker(2*time.Second, 2)
	check.RunInterval(ctx, 1*time.Second, healthChecker)

	healthPortInt, err := docker.GetFreePort()
	if err != nil {
		t.Error(err)
		return
	}
	healthPort := strconv.Itoa(healthPortInt)
	health.StartEndpoint(ctx, healthPort, healthChecker)

	healthChecker2 := NewHealthChecker(1*time.Second, 2)
	check.RunInterval(ctx, 1*time.Hour, healthChecker2)

	healthPortInt2, err := docker.GetFreePort()
	if err != nil {
		t.Error(err)
		return
	}
	healthPort2 := strconv.Itoa(healthPortInt2)
	health.StartEndpoint(ctx, healthPort2, healthChecker2)

	time.Sleep(5 * time.Second)
	t.Run("check expect healthy", testCheckHealth(healthPort, true))
	t.Run("check expect missing interval call", testCheckHealth(healthPort2, false))
}

func TestHealthCheckMissingVerne(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	wg := &sync.WaitGroup{}
	defer wg.Wait()
	defer cancel()

	loggerMock := mocks.Logger()
	stateMock := mocks.State()
	iotMock := mocks.Devices()

	freePort, err := docker.GetFreePort()
	if err != nil {
		t.Error(err)
		return
	}

	check := ConnectionCheck{
		Logger:                     loggerMock,
		LoggerState:                stateMock,
		Verne:                      vernemq.New("http://localhost:" + strconv.Itoa(freePort)),
		Devices:                    iotMock,
		TokenGen:                   mocks.TokenGen,
		SubscriptionTopicGenerator: topicgenerator.Known["senergy"],
		BatchSize:                  2,
		HandledProtocols:           map[string]bool{"test-protocol": true},
		Debug:                      true,
	}

	t.Run("create device type 1", testCreateDeviceType(iotMock, "dt1", []model.Service{
		{
			LocalId:     "sl1",
			ProtocolId:  "test-protocol",
			FunctionIds: []string{model.CONTROLLING_FUNCTION_PREFIX + "f1"},
		},
	}))

	t.Run("create device", testCreateDevice(iotMock, stateMock, "device", "dt1", true))

	healthChecker := NewHealthChecker(2*time.Hour, 2) //irrelevant time
	check.RunInterval(ctx, 1*time.Second, healthChecker)

	healthPortInt, err := docker.GetFreePort()
	if err != nil {
		t.Error(err)
		return
	}
	healthPort := strconv.Itoa(healthPortInt)

	health.StartEndpoint(ctx, healthPort, healthChecker)

	time.Sleep(5 * time.Second)
	t.Run("check", testCheckHealth(healthPort, false))
}

func testCheckHealth(port string, expectHealthy bool) func(t *testing.T) {
	return func(t *testing.T) {
		resp, err := http.Get("http://localhost:" + port)
		if err != nil {
			t.Error(err)
			return
		}
		info, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			t.Error(err)
			return
		}
		t.Log(string(info))
		if (resp.StatusCode == 200) != expectHealthy {
			t.Error(resp.StatusCode)
			return
		}
	}
}
