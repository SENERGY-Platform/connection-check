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

package health

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"
)

func StartEndpoint(ctx context.Context, port string, checkable Checkable) {
	log.Println("start health-check api on " + port)
	server := &http.Server{
		Addr:              ":" + port,
		Handler:           getHealthCheckEndpoint(checkable),
		WriteTimeout:      10 * time.Second,
		ReadTimeout:       2 * time.Second,
		ReadHeaderTimeout: 2 * time.Second,
	}
	go func() {
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			log.Println("ERROR: server error", err)
			log.Fatal(err)
		}
	}()
	go func() {
		<-ctx.Done()
		log.Println("DEBUG: health-check shutdown", server.Shutdown(context.Background()))
	}()
	return
}

func getHealthCheckEndpoint(checkable Checkable) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		ok, info := checkable.Check()
		if ok {
			json.NewEncoder(writer).Encode(info)
		} else {
			temp, _ := json.Marshal(info)
			http.Error(writer, string(temp), 500)
		}
	}
}
