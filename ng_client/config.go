package main

import (
	"encoding/json"
	"log"
	"os"
)

type Configuration struct {
    Server string
    Username string    
    Status string    
}

const fname = "config.json"

func GetConfiguration() Configuration {
	configuration := Configuration{}
    file, err := os.OpenFile(fname, os.O_RDWR, 0600)
    defer file.Close()	
    if err != nil {
        if os.IsNotExist(err) {            
        	configuration = Configuration{Server: defaultAddress,Username: defaultName,Status: "Hey! Im new in console" }

        	filecontent, _ := json.MarshalIndent(configuration, "", " ")            
            f, err := os.Create(fname)
    		_, err = f.Write(filecontent)

    		if err != nil {
			  log.Println("error creating file :", err)
			}

    		f.Sync()    		
        } else {
        	log.Println("error opening file :", err)
        }        
    } else {
    	decoder := json.NewDecoder(file)		
		err := decoder.Decode(&configuration)
		if err != nil {
		  log.Println("error decoding file:", err)
		}
    }
	return configuration
}

func SetConfiguration(Server string, Username string, Status string) Configuration {	
	configuration := Configuration{Server: Server,Username: Username,Status: Status }	
    file, err := os.OpenFile(fname, os.O_RDWR|os.O_CREATE, 0600)
    defer file.Close()	
    if err != nil {
        log.Println("error opening file :", err)     
    } else {
    	file.Truncate(0)
		file.Seek(0,0)

    	filecontent, _ := json.MarshalIndent(configuration, "", " ")
    	_, err = file.Write(filecontent)

    	file.Sync()

    	if err != nil {
		  log.Println("error writing file :", err)
		}
    }
	return configuration	
}