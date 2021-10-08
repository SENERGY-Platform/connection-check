/*
 * Copyright 2021 InfAI (CC SES)
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
	"log"
	"strconv"
	"strings"
)

func ParseAssignmentId(assignmentId string) (assignmentIndex int, err error) {
	parts := strings.Split(assignmentId, "-")
	last := parts[len(parts)-1]
	return strconv.Atoi(last)
}

func IsAssignedBatch(batchSize int, offset int, scale int, assignmentIndex int) bool {
	if scale < 1 {
		log.Println("WARNING: configured scale < 1 --> use scale 1")
		scale = 1
	}
	batchIndex := offset / batchSize
	return (batchIndex % scale) == assignmentIndex
}
