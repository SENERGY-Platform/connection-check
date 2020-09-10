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

package main

import (
	connectioncheck "connection-check/pkg"
	"connection-check/pkg/configuration"
	"connection-check/pkg/health"
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	time.Sleep(5 * time.Second) //wait for routing tables in cluster

	confLocation := flag.String("config", "config.json", "configuration file")
	flag.Parse()

	config, err := configuration.Load(*confLocation)
	if err != nil {
		log.Fatal("ERROR: unable to load config ", err)
	}

	check, err := connectioncheck.New(config)
	if err != nil {
		log.Fatal("ERROR: unable to init connectioncheck ", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	healthChecker := connectioncheck.NewHealthChecker(time.Duration(config.IntervalSeconds)*time.Second*2, config.HealthErrorLimit)
	check.RunInterval(ctx, time.Duration(config.IntervalSeconds)*time.Second, healthChecker)
	health.StartEndpoint(ctx, config.HealthPort, healthChecker)

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)
	sig := <-shutdown
	log.Println("received shutdown signal", sig)
}
