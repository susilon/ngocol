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
	"bufio"
	"fmt"
	"os"	
	"log"
	"io"

	"context"		
	"time"
	"encoding/json"

	"google.golang.org/grpc"
	pb "google.golang.org/grpc/examples/ngocol/ngocol"  
)

const (	
	defaultAddress = "localhost:50051"
	defaultName = "Guest"
	defaultDestination = "all"
)

func main() {		
	fmt.Println("NgoCol : Ngobrol di Console")
	fmt.Println("-------- /q to quit--------")

	//filename is the path to the json config file
	configuration := GetConfiguration()	
	fmt.Println("Hello", configuration.Username)	
		
	destination := defaultDestination	
	if len(os.Args) > 1 {		
		switch os.Args[1] {			
			case "-setup" :
				configuration.Username = "Guest"							
			default :
				destination = os.Args[1]
		}
	}

	var paramRequest bool
	var paramType string	
	var c pb.NgocolClient

	if (configuration.Username == "Guest") {
		// Starting Configuration Wizard
		paramRequest = true
		paramType = "un"
		fmt.Println("Please Enter Server Address (servername:port)")		
	} else {		
		// Set up a connection to the server.
		fmt.Printf("Connecting To Server: %s", configuration.Server)
		conn, err := grpc.Dial(configuration.Server, grpc.WithInsecure())
		if err != nil {
			fmt.Println("did not connect: %v", err)		
		}
		defer conn.Close()
		c = pb.NewNgocolClient(conn)	

		fmt.Println("\rConnected To Server:", configuration.Server, "  ")	
		// Start Bidirectional gRPC
		go waitchat(c, configuration.Username, destination, configuration.Status)
	}
		
	scanner := bufio.NewScanner(os.Stdin)	
	for scanner.Scan() {			
		if (paramRequest) {
			if (paramType == "q" && scanner.Text() != "N" && scanner.Text() != "n"){
				ctxs, cancels := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancels()		    	
				st, err := c.SendText(ctxs, &pb.MsgData{User: configuration.Username, Message: "quit", Destination: "srv"})
				if err != nil {
					fmt.Println("could not send: %v", err)						
				}				

				_responseMessage := pb.ResponseMessage{}
				json.Unmarshal([]byte(st.Message), &_responseMessage)				
				if (_responseMessage.Status == "false") {
					fmt.Println("Error : " + _responseMessage.Data)
				}						
				paramRequest = false				
				return
			} 

			switch paramType {			
			case "un" :
				configuration.Server = scanner.Text()
				paramType = "ur"
				fmt.Println("Please Enter Your Name")					
			case "ur" :
				configuration.Username = scanner.Text()
				paramType = "us"
				fmt.Println("Please Enter Your Status")					
			case "us" :
				configuration.Status = scanner.Text()

				SetConfiguration(configuration.Server, configuration.Username, configuration.Status)
				paramType = ""
				paramRequest = false

				fmt.Println("Thank You")	
				fmt.Println("Configuration Finished : ", configuration)				
				
				fmt.Printf("Connecting To Server:", configuration.Server)
				conn, err := grpc.Dial(configuration.Server, grpc.WithInsecure())
				if err != nil {
					fmt.Println("did not connect: %v", err)		
				}
				defer conn.Close()
				c = pb.NewNgocolClient(conn)	
				fmt.Println("\rConnected To Server:", configuration.Server, "  ")	
				// Start Bidirectional gRPC
				go waitchat(c, configuration.Username, destination, configuration.Status)
			}						
		} else {
			switch scanner.Text() {
		    case "/q":
		        paramRequest = true
		        paramType = "q"
				fmt.Println("Quit (Y)/N?")		
			case "/c":
		        paramRequest = true
				paramType = "un"
				fmt.Println("Please Enter Server Address (servername:port)")
			case "/l":
				fmt.Println("-------------")
		        fmt.Println("Online Users:")
		        users := GetUserList(c, configuration.Username, configuration.Status)		        
		        for _, user := range users.Users {
			        fmt.Printf("Name: %v, Status: %v\n", user.Username, user.Status)        
			    }		    
			    fmt.Println("-------------")
		    default:
		    	ctxs, cancels := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancels()		    	
				st, err := c.SendText(ctxs, &pb.MsgData{User: configuration.Username, Message: scanner.Text(), Destination: destination})
				if err != nil {
					fmt.Println("could not send: %v", err)						
				}				

				_responseMessage := pb.ResponseMessage{}
				json.Unmarshal([]byte(st.Message), &_responseMessage)				
				if (_responseMessage.Status == "false") {
					fmt.Println("Error : " + _responseMessage.Data)
				}
				paramRequest = false
		    }
		}		
	}
}

func waitchat(c pb.NgocolClient, name string, destination string, status string) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()	

	fmt.Printf("Starting Listener")
	stream, err := c.WaitText(ctx)
	if err != nil {
		log.Fatalf("%v.WaitText(_) = _, %v", c, err)
	}
	waitc := make(chan struct{})
	fmt.Println("\rReady            ")
	go func() {
		// infinity loop started, waiting stream from server
		for {			
			in, err := stream.Recv()
			if err == io.EOF {
				// read done.
				close(waitc)
				log.Fatalf("Stream closed : %v", err)
				return 
			}
			if err != nil {
				log.Fatalf("Failed to receive a message : %v", err)
				return
			}
						
			if (in.User != "srv") {
				log.Printf(":%s> %s", in.User, in.Message)
			} else {
				if (in.Message == "quit") {
					log.Println("Quit Program")
					return
				}				
			}	
		}
	}()		
	//data := "{\"u\":\"" + name + "\",\"m\":\"" + status + "\",\"d\":\"" + destination + "\"}"
	if err := stream.Send(&pb.MsgData{User: name, Message: status, Destination: destination}); err != nil {
		log.Fatalf("Failed to send a note: %v", err)
	}	
	stream.CloseSend()
	<-waitc	
	log.Fatalf("Stream closed at last : %v", err)
}