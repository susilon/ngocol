/*
 *
 * Copyright 2015 gRPC authors.
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

//protoc -I ngocol/ ngocol/ngocol.proto --go_out=plugins=grpc:ngocol

package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"time"
	"io"	
	"strings"
	//"encoding/json"
	"container/list"

	"google.golang.org/grpc"
	pb "google.golang.org/grpc/examples/ngocol/ngocol"
)

type server struct{	
	users list.List
	listmessage list.List
	ischeck bool
	user pb.User
}

const (
	defaultPort = ":50051"
)

func (s *server) SendText(ctx context.Context, in *pb.MsgData) (*pb.MsgData, error) {
	// new messages arrived
	fmt.Printf("Messages from %s to %s: %s\n", in.User, in.Destination, in.Message)
	if (in.Message == "quit" && in.Destination == "srv") {
		msg := pb.MsgData{User: "Server", Message: in.User + " is Quit", Destination: "all"}
		// enqueue messages
		s.listmessage.PushBack(&msg)
		// remove user
		for usr := s.users.Front(); usr != nil; usr = usr.Next() {
			if user, ok := usr.Value.(pb.User); ok {	        	
				if (strings.ToLower(in.User) == user.Username) {
					// remove user from user lists
					log.Printf("Removing %s", user.Username)
					s.users.Remove(usr)
				}			        
			}
		}	
	} else {
		// enqueue messages
		s.listmessage.PushBack(in)
	}	
	// response to client	
	return &pb.MsgData{User:"srv",Message:"{\"Status\" : \"true\", \"Data\":\"" + in.Message + "\"}",Destination:in.User}, nil
}

func (s *server) ListUser(ctx context.Context, in *pb.User) (*pb.UserList, error) {	
	// user request list of user
	log.Printf("User List Request From: %s", in.Username)

	var _userlist pb.UserList		
	
	for usr := s.users.Front(); usr != nil; usr = usr.Next() {				
		useritem := usr.Value.(pb.User)
		_userlist.Users = append(_userlist.Users, &useritem)
	}

	log.Printf("Return: %v", &_userlist)
	return &_userlist, nil
}

func UserExist(U string, L list.List) bool {
	userexists := false
	for usr := L.Front(); usr != nil; usr = usr.Next() {
		if user, ok := usr.Value.(pb.User); ok {	        	
			if (U == user.Username) {
				userexists = true
			}			        
		}
	}
	return userexists
}

func (s *server) WaitText(stream pb.Ngocol_WaitTextServer) error {	
	in, err := stream.Recv()
	if err == io.EOF {
		return err
	}
	if err != nil {
		return err
	}
	
	// new user is connected
	fmt.Printf("New User Online: %s\n", in.User)
	//receivedmsg := msgdata{}
	//json.Unmarshal([]byte(in.Data), &receivedmsg)
	U := strings.ToLower(in.User)
	D := strings.ToLower(in.Destination)
	s.user = pb.User{Username:U, Status:in.Message}
	
	// add new user to user lists
	if (!UserExist(U, s.users)) {		
		s.users.PushBack(s.user)			
	}	

	// broadcast info
	msg := pb.MsgData{User: "Server", Message: in.User + " is Online", Destination: "all"}
	// enqueue messages
	s.listmessage.PushBack(&msg)
	
	// infinity loop started
	for {
		// go need this or hang
		time.Sleep(1 * time.Nanosecond)

		for s.listmessage.Len() > 0 {			
	        // print first message
	        e := s.listmessage.Front()
	        
	        if str, ok := e.Value.(*pb.MsgData); ok {
			    if (strings.ToLower(str.Destination) == U && strings.ToLower(str.User) == D) {
			    	log.Printf("Sending %s, from %s to %s", str.Message, str.User, str.Destination)					
					
					if err := stream.Send(&pb.MsgData{User: str.User, Message: str.Message, Destination: str.Destination}); err != nil {						
						// error while sending steram to client, user maybe offline
						log.Printf("Error : %s, User : %s is Offline", err, U)
						for usr := s.users.Front(); usr != nil; usr = usr.Next() {
							if user, ok := usr.Value.(*pb.User); ok {	        	
								if (U == user.Username) {
									// remove user from user lists
									log.Printf("Removing %s", &user.Username)
									s.users.Remove(usr)
								}			        
							}
						}						

						log.Printf("Error %s", err)
						return err
					} else {
						// dequeue message
		        		s.listmessage.Remove(e)
					}
				} else if (str.Destination == "all" && D == "all" && strings.ToLower(str.User) != U) {
					// sending to all connected client
					log.Printf("Sending %s, from %s to all, receiver %s", str.Message, str.User, U)					
					if err := stream.Send(&pb.MsgData{User: str.User, Message: str.Message, Destination: str.Destination}); err != nil {
						// error while sending steram to client, user maybe offline
						log.Printf("Error : %s, User : %s is Offline", err, U)
						for usr := s.users.Front(); usr != nil; usr = usr.Next() {
							if user, ok := usr.Value.(pb.User); ok {	        	
								if (U == user.Username) {
									// remove user from user lists
									log.Printf("Removing %s", user.Username)
									s.users.Remove(usr)
								}			        
							}
						}						

						return err
					} else {
						// dequeue message
		        		s.listmessage.Remove(e)
					}
				} else if (str.Destination == "all" && strings.ToLower(str.User) == "srv") {
					// sending to all connected client
					log.Printf("Broadcast %s, from %s to all, receiver %s", str.Message, str.User, U)					
					if err := stream.Send(&pb.MsgData{User: str.User, Message: str.Message, Destination: str.Destination}); err != nil {
						// error while sending steram to client, user maybe offline
						log.Printf("Error : %s, User : %s is Offline", err, U)
						for usr := s.users.Front(); usr != nil; usr = usr.Next() {
							if user, ok := usr.Value.(pb.User); ok {	        	
								if (U == user.Username) {
									// remove user from user lists
									log.Printf("Removing %s", user.Username)
									s.users.Remove(usr)
								}			        
							}
						}						

						return err
					} else {
						// dequeue message
		        		s.listmessage.Remove(e)
					}
				} 
			} else {
				log.Printf("Not OK: %s", e )
				s.listmessage.Remove(e)
				return nil
			}
		}	
	}	
}

func ServerTick() {
	for {	
		time.Sleep(1 * time.Second)
		fmt.Printf("\rCheck %s", time.Now())
	}
}

func main() {	
	port := defaultPort
	if len(os.Args) > 1 {
		port = ":" + os.Args[1]
	} 
	println("Listening at " + port)
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}	

	s := grpc.NewServer()	
	pb.RegisterNgocolServer(s, &server{})	
	
	go ServerTick() // update screen every second, just to show the app still running

	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
