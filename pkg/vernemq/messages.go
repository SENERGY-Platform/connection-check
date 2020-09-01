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

type Client struct {
	Id   string `json:"client_id"`
	User string `json:"user"`
}

type ClientWrapper struct {
	Table []Client `json:"table"`
}

type Subscription struct {
	ClientId string `json:"client_id"`
	User     string `json:"user"`
	Topic    string `json:"topic"`
}

type SubscriptionWrapper struct {
	Table []Subscription `json:"table"`
}
