/*
 *
 * Copyright 2019 susilon.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */
package main

import (
	"context"
	"log"	
	"time"	
	//"encoding/json"

	//"google.golang.org/grpc"
	pb "google.golang.org/grpc/examples/ngocol/ngocol"
)

func GetUserList(c pb.NgocolClient, name string, status string) *pb.UserList {	
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	
	rg, err := c.ListUser(ctx, &pb.User{Username: name, Status: status})
	if err != nil {
		log.Fatalf("could not get user list: %v", err)
	}	
	
	return rg
}
