package main

import (
	"bhasherbel/peeruptime/types"
	"fmt"
	"os"
	"os/signal"
)

var config = types.NewConfig()
var hosts = types.NewHosts()
var server = &Server{Hosts: hosts, Config: config.Server}
var client = &Client{Hosts: hosts, Config: config.Client}

func main() {
	hosts.AppendNew(config.Client.DiscoveryURL)

	if config.Server.Enable {
		go server.Start()
		fmt.Println("Server listening on " + config.Server.Ip + ":" + config.Server.Port)
	}

	if config.Client.Enable {
		go client.Start()
		fmt.Println("Check loop started")
	}

	sigint := make(chan os.Signal, 1)
	signal.Notify(sigint, os.Interrupt)

	<-sigint

	fmt.Println("Received SIGINT signal. Shuting down...")
	os.Exit(0)
}
