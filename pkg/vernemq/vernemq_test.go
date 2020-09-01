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
	"connection-check/pkg/test/docker"
	"context"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/ory/dockertest/v3"
	"reflect"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestGetOnlineSubscriptions(t *testing.T) {
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

	t.Run("create test client 1", testStartTestClient(ctx, wg, brokerUrl, "client1", []string{"topic1", "topic2"}))
	t.Run("create test client 2", testStartTestClient(ctx, wg, brokerUrl, "client2", []string{"topic2", "topic3"}))

	t.Run("read online clients", testReadOnlineClients(managementUrl, []Client{
		{
			Id:   "client1",
			User: "test",
		},
		{
			Id:   "client2",
			User: "test",
		},
	}))

	t.Run("read online subscriptions", testReadOnlineSubscriptions(managementUrl, []Subscription{
		{
			ClientId: "client1",
			User:     "test",
			Topic:    "topic1",
		},
		{
			ClientId: "client1",
			User:     "test",
			Topic:    "topic2",
		},
		{
			ClientId: "client2",
			User:     "test",
			Topic:    "topic2",
		},
		{
			ClientId: "client2",
			User:     "test",
			Topic:    "topic3",
		},
	}))

	offlineCtx, offlineCancel := context.WithCancel(ctx)
	t.Run("create test offline client 3", testStartTestClient(offlineCtx, &sync.WaitGroup{}, brokerUrl, "client3", []string{"topic3", "topic4"}))
	offlineCancel()
	time.Sleep(1 * time.Second)

	t.Run("read online clients", testReadOnlineClients(managementUrl, []Client{
		{
			Id:   "client1",
			User: "test",
		},
		{
			Id:   "client2",
			User: "test",
		},
	}))

	t.Run("read online subscriptions", testReadOnlineSubscriptions(managementUrl, []Subscription{
		{
			ClientId: "client1",
			User:     "test",
			Topic:    "topic1",
		},
		{
			ClientId: "client1",
			User:     "test",
			Topic:    "topic2",
		},
		{
			ClientId: "client2",
			User:     "test",
			Topic:    "topic2",
		},
		{
			ClientId: "client2",
			User:     "test",
			Topic:    "topic3",
		},
	}))
}

func TestCheckOnlineSubscription(t *testing.T) {
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

	t.Run("create test client with placeholder", testStartTestClient(ctx, wg, brokerUrl, "placeholder", []string{"with/placeholder/#"}))
	t.Run("create test client 1", testStartTestClient(ctx, wg, brokerUrl, "client1", []string{"topic1", "topic2"}))
	t.Run("create test client 2", testStartTestClient(ctx, wg, brokerUrl, "client2", []string{"topic2", "topic3"}))
	t.Run("create test client foo", testStartTestClient(ctx, wg, brokerUrl, "client3", []string{"foo/bar", "foo/bar/batz"}))
	t.Run("create test client foo", testStartTestClient(ctx, wg, brokerUrl, "senergy", []string{"urn:senergy:foo-bar-batz", "urn:senergy:foo-bar-batz/localService"}))

	offlineCtx, offlineCancel := context.WithCancel(ctx)
	t.Run("create test offline client 3", testStartTestClient(offlineCtx, &sync.WaitGroup{}, brokerUrl, "client4", []string{"topic3", "topic4"}))
	offlineCancel()
	time.Sleep(1 * time.Second)

	t.Run(testCheckOnlineSubscription(managementUrl, "with/placeholder/#", true))
	t.Run(testCheckOnlineSubscription(managementUrl, "with/placeholder", false))
	t.Run(testCheckOnlineSubscription(managementUrl, "with/placeholder/foo", false))
	t.Run(testCheckOnlineSubscription(managementUrl, "topic1", true))
	t.Run(testCheckOnlineSubscription(managementUrl, "topic2", true))
	t.Run(testCheckOnlineSubscription(managementUrl, "topic3", true))
	t.Run(testCheckOnlineSubscription(managementUrl, "urn:senergy:foo-bar-batz", true))
	t.Run(testCheckOnlineSubscription(managementUrl, "urn:senergy:foo-bar-batz/localService", true))
	t.Run(testCheckOnlineSubscription(managementUrl, "topic4", false))
	t.Run(testCheckOnlineSubscription(managementUrl, "foo", false))
	t.Run(testCheckOnlineSubscription(managementUrl, "foo/bar", true))
	t.Run(testCheckOnlineSubscription(managementUrl, "foo/bar/batz", true))
	t.Run(testCheckOnlineSubscription(managementUrl, "unknown", false))
	t.Run(testCheckOnlineSubscription(managementUrl, "foo/unknown", false))
	t.Run(testCheckOnlineSubscription(managementUrl, "unknown/foo", false))
}

func TestCheckOnlineClient(t *testing.T) {
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

	t.Run("create test client 1", testStartTestClient(ctx, wg, brokerUrl, "client1", []string{"topic1", "topic2"}))
	t.Run("create test client 2", testStartTestClient(ctx, wg, brokerUrl, "client2", []string{"topic2", "topic3"}))
	t.Run("create test client foo", testStartTestClient(ctx, wg, brokerUrl, "client3", []string{"foo/bar", "foo/bar/batz"}))
	t.Run("create test client foo", testStartTestClient(ctx, wg, brokerUrl, "uuid:senergy:foo-bar-batz", []string{"urn"}))

	offlineCtx, offlineCancel := context.WithCancel(ctx)
	t.Run("create test offline client 3", testStartTestClient(offlineCtx, &sync.WaitGroup{}, brokerUrl, "client4", []string{"topic3", "topic4"}))
	offlineCancel()
	time.Sleep(1 * time.Second)

	t.Run(testCheckOnlineClient(managementUrl, "client1", true))
	t.Run(testCheckOnlineClient(managementUrl, "client2", true))
	t.Run(testCheckOnlineClient(managementUrl, "client3", true))
	t.Run(testCheckOnlineClient(managementUrl, "client4", false))
	t.Run(testCheckOnlineClient(managementUrl, "unknown", false))
	t.Run(testCheckOnlineClient(managementUrl, "uuid:senergy:foo-bar-batz", true))
	t.Run(testCheckOnlineClient(managementUrl, "uuid:senergy:foo-bar-batz-unknown", false))
}

func testCheckOnlineClient(url string, clientId string, expected bool) (string, func(t *testing.T)) {
	return strings.Replace(clientId, "/", " ", -1), func(t *testing.T) {
		api := &VernemqManagementApi{
			Url:             url,
			NodeResultLimit: 10000,
		}
		result, err := api.CheckOnlineClient(clientId)
		if err != nil {
			t.Error(err)
			return
		}
		if result != expected {
			t.Error(result, expected)
		}
	}
}

func testCheckOnlineSubscription(url string, topic string, expected bool) (string, func(t *testing.T)) {
	return strings.Replace(topic, "/", " ", -1), func(t *testing.T) {
		api := &VernemqManagementApi{
			Url:             url,
			NodeResultLimit: 10000,
		}
		result, err := api.CheckOnlineSubscription(topic)
		if err != nil {
			t.Error(err)
			return
		}
		if result != expected {
			t.Error(result, expected)
		}
	}
}

func testReadOnlineClients(url string, expected []Client) func(t *testing.T) {
	return func(t *testing.T) {
		api := &VernemqManagementApi{
			Url:             url,
			NodeResultLimit: 10000,
		}
		result, err := api.GetOnlineClients()
		if err != nil {
			t.Error(err)
			return
		}
		sort.SliceStable(expected, func(i, j int) bool {
			return expected[i].Id < expected[j].Id
		})
		sort.SliceStable(result, func(i, j int) bool {
			return result[i].Id < result[j].Id
		})

		if !reflect.DeepEqual(result, expected) {
			t.Error(result, expected)
			return
		}
	}
}

func testReadOnlineSubscriptions(url string, expected []Subscription) func(t *testing.T) {
	return func(t *testing.T) {
		api := &VernemqManagementApi{
			Url:             url,
			NodeResultLimit: 10000,
		}
		result, err := api.GetOnlineSubscriptions()
		if err != nil {
			t.Error(err)
			return
		}
		sort.SliceStable(expected, func(i, j int) bool {
			return expected[i].Topic < expected[j].Topic
		})
		sort.SliceStable(expected, func(i, j int) bool {
			return expected[i].ClientId < expected[j].ClientId
		})
		sort.SliceStable(result, func(i, j int) bool {
			return result[i].Topic < result[j].Topic
		})
		sort.SliceStable(result, func(i, j int) bool {
			return result[i].ClientId < result[j].ClientId
		})

		if !reflect.DeepEqual(result, expected) {
			t.Error(result, expected)
			return
		}
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
