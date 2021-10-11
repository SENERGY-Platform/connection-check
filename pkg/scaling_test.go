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
	"fmt"
)

func ExampleIsAssignedBatch_1_0() {
	batchSize := 500
	scale := 1
	assignmentIndex, err := ParseAssignmentId("check-senergy-0")
	if err != nil {
		fmt.Println(err)
		return
	}
	for i := 0; i < 10000; i = i + batchSize {
		fmt.Println(IsAssignedBatch(batchSize, i, scale, assignmentIndex))
	}

	//output:
	//true
	//true
	//true
	//true
	//true
	//true
	//true
	//true
	//true
	//true
	//true
	//true
	//true
	//true
	//true
	//true
	//true
	//true
	//true
	//true
}

func ExampleIsAssignedBatch_2_0() {
	batchSize := 500
	scale := 2
	assignmentIndex, err := ParseAssignmentId("check-senergy-0")
	if err != nil {
		fmt.Println(err)
		return
	}
	for i := 0; i < 10000; i = i + batchSize {
		fmt.Println(IsAssignedBatch(batchSize, i, scale, assignmentIndex))
	}

	//output:
	//true
	//false
	//true
	//false
	//true
	//false
	//true
	//false
	//true
	//false
	//true
	//false
	//true
	//false
	//true
	//false
	//true
	//false
	//true
	//false
}

func ExampleIsAssignedBatch_2_1() {
	batchSize := 500
	scale := 2
	assignmentIndex, err := ParseAssignmentId("check-senergy-1")
	if err != nil {
		fmt.Println(err)
		return
	}
	for i := 0; i < 10000; i = i + batchSize {
		fmt.Println(IsAssignedBatch(batchSize, i, scale, assignmentIndex))
	}

	//output:
	//false
	//true
	//false
	//true
	//false
	//true
	//false
	//true
	//false
	//true
	//false
	//true
	//false
	//true
	//false
	//true
	//false
	//true
	//false
	//true
}

func ExampleIsAssignedBatch_3_0() {
	batchSize := 500
	scale := 3
	assignmentIndex, err := ParseAssignmentId("check-senergy-0")
	if err != nil {
		fmt.Println(err)
		return
	}
	for i := 0; i < 10000; i = i + batchSize {
		fmt.Println(IsAssignedBatch(batchSize, i, scale, assignmentIndex))
	}

	//output:
	//true
	//false
	//false
	//true
	//false
	//false
	//true
	//false
	//false
	//true
	//false
	//false
	//true
	//false
	//false
	//true
	//false
	//false
	//true
	//false
}