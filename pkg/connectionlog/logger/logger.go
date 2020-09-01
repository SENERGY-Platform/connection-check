/*
 * Copyright 2019 InfAI (CC SES)
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

package logger

import (
	"connection-check/pkg/connectionlog/logger/kafka"
	"encoding/json"
	"time"
)

func New(zk string, sync bool, idempotent bool, deviceLogTopic string, hubLogTopic string) (logger *Logger, err error) {
	producer, err := kafka.PrepareProducer(zk, sync, idempotent)
	if err != nil {
		return logger, err
	}
	return &Logger{producer: producer, deviceLogTopic: deviceLogTopic, hubLogTopic: hubLogTopic}, nil
}

type Logger struct {
	producer       kafka.ProducerInterface
	deviceLogTopic string
	hubLogTopic    string
}

func (this *Logger) LogDeviceDisconnect(id string) error {
	b, err := json.Marshal(DeviceLog{
		Connected: false,
		Id:        id,
		Time:      time.Now(),
	})
	if err != nil {
		return err
	}
	return this.producer.ProduceWithKey(this.deviceLogTopic, string(b), id)
}

func (this *Logger) LogDeviceConnect(id string) error {
	b, err := json.Marshal(DeviceLog{
		Connected: true,
		Id:        id,
		Time:      time.Now(),
	})
	if err != nil {
		return err
	}
	return this.producer.ProduceWithKey(this.deviceLogTopic, string(b), id)
}

func (this *Logger) LogHubConnect(id string) error {
	b, err := json.Marshal(HubLog{
		Connected: true,
		Id:        id,
		Time:      time.Now(),
	})
	if err != nil {
		return err
	}
	return this.producer.ProduceWithKey(this.hubLogTopic, string(b), id)
}

func (this *Logger) LogHubDisconnect(id string) error {
	b, err := json.Marshal(HubLog{
		Connected: false,
		Id:        id,
		Time:      time.Now(),
	})
	if err != nil {
		return err
	}
	return this.producer.ProduceWithKey(this.hubLogTopic, string(b), id)
}

func (this *Logger) Close() {
	this.producer.Close()
}
