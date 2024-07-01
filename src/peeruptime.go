package main

import (
	"bhasherbel/peeruptime/types"
	"bhasherbel/peeruptime/utils"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"time"
)

// Structures

type Config struct {
	DiscoveryLimit      int
	Ip                  string
	Port                string
	URL                 string
	DiscoveryURL        string
	MemoryScoreFactor   int
	VariableScoreFactor int
	WaitingTime         int // ms
	Server              bool
	Client              bool
}

type Response struct {
	Status      string
	Config      *SharedConfig
	Discoveries []string
}

type SharedConfig struct {
}

type StatusRequest struct {
	Discovery          bool
	DiscoveryLimit     int
	Discoverable       bool
	DiscoverableURL    string
	DiscoverableConfig *SharedConfig
}

// Global variables

var config = Config{
	DiscoveryLimit:      utils.IntOrDefault(os.Getenv("PEER_DISCOVERY_LIMIT"), 5),
	Ip:                  utils.StringOrDefault(os.Getenv("PEER_IP"), "0.0.0.0"),
	Port:                utils.StringOrDefault(os.Getenv("PEER_PORT"), "8080"),
	URL:                 utils.StringOrDefault(os.Getenv("PEER_URL"), "http://127.0.0.1:8080"),
	DiscoveryURL:        utils.StringOrDefault(os.Getenv("PEER_DISCOVERY_URL"), "http://localhost:8081"),
	MemoryScoreFactor:   utils.IntOrDefault(os.Getenv("PEER_MEMORY_SCORE_FACTOR"), 10),
	VariableScoreFactor: utils.IntOrDefault(os.Getenv("PEER_VARIABLE_SCORE_FACTOR"), 1000),
	WaitingTime:         utils.IntOrDefault(os.Getenv("PEER_WAITING_TIME"), 1000),
	Server:              utils.StringOrDefault(os.Getenv("PEER_SERVER"), "true") == "true",
	Client:              utils.StringOrDefault(os.Getenv("PEER_CLIENT"), "true") == "true",
}

var hosts = types.NewHosts()

// Main

func main() {
	hosts.AppendNew(config.DiscoveryURL)

	if config.Server {
		http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("OK"))
		})

		http.HandleFunc("/api/status", statusHandler)

		go http.ListenAndServe(config.Ip+":"+config.Port, nil)
		fmt.Println("Server started on " + config.Ip + ":" + config.Port)
	}

	if config.Client {
		go loopCheck()
		fmt.Println("Check loop started")
	}

	sigint := make(chan os.Signal, 1)
	signal.Notify(sigint, os.Interrupt)

	<-sigint

	fmt.Println("Received SIGINT signal. Shuting down...")
	os.Exit(0)
}

func statusHandler(w http.ResponseWriter, r *http.Request) {
	request := &StatusRequest{}

	json.NewDecoder(r.Body).Decode(request)

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Request-Time", time.Now().UTC().Format(time.RFC3339Nano))

	response := Response{
		Status:      "OK",
		Config:      &SharedConfig{},
		Discoveries: *hosts.Keys,
	}

	json.NewEncoder(w).Encode(response)

	go func() {
		if request.Discoverable && request.DiscoverableURL != "" && request.DiscoverableURL != config.URL {
			if _, ok := hosts.Get(request.DiscoverableURL); !ok {
				fmt.Println("Discovered new host: " + request.DiscoverableURL)
				hosts.AppendNew(request.DiscoverableURL)
			}
		}
	}()
}

func loopCheck() {
	for {
		check()
		time.Sleep(time.Duration(config.WaitingTime) * time.Millisecond)
	}
}

func check() {
	item := hosts.Peek()
	url := item.Value
	host, _ := hosts.Get(url)

	fmt.Printf("Checking %v ... ", url)

	check := checkHost(host)
	host.Checks.Append(check, float64(config.MemoryScoreFactor))

	if check.Success {
		fmt.Printf("success (%vms, %vms, %vms)\n", check.PingDelay, check.PongDelay, check.LocalDelay)
	} else {
		fmt.Println("Failed")
	}

	newPriority := int(host.Checks.AmortizedScore())

	hosts.UpdatePriority(item, newPriority)
}

func checkHost(host *types.Host) types.Check {
	before := time.Now().UTC()

	request := StatusRequest{
		Discovery:          true,
		DiscoveryLimit:     config.DiscoveryLimit,
		Discoverable:       true,
		DiscoverableURL:    config.URL,
		DiscoverableConfig: &SharedConfig{},
	}

	jsonData, err := json.Marshal(request)
	if err != nil {
		fmt.Printf("%v ", err)
		return types.Check{Time: before, Success: false}
	}

	resp, err := http.Post(host.URL+"/api/status", "application/json", bytes.NewBuffer(jsonData))
	after := time.Now().UTC()
	if err != nil {
		fmt.Printf("%v ", err)
		return types.Check{Time: before, Success: false}
	}
	defer resp.Body.Close()

	pingTime, err := time.Parse(time.RFC3339, resp.Header.Get("X-Request-Time"))
	if err != nil {
		fmt.Printf("%v ", err)
		return types.Check{Time: before, Success: false}
	}

	response := &Response{}

	json.NewDecoder(resp.Body).Decode(response)

	for _, discovery := range response.Discoveries {
		if discovery == config.URL {
			continue
		}
		if _, ok := hosts.Get(discovery); !ok {
			fmt.Println("\nDiscovered new host: " + discovery)
			hosts.AppendNew(discovery)
		}
	}

	return types.Check{
		Time:       before,
		PingDelay:  float64(pingTime.Sub(before).Microseconds()) / 1000.,
		PongDelay:  float64(after.Sub(pingTime).Microseconds()) / 1000.,
		LocalDelay: float64(after.Sub(before).Microseconds()) / 1000.,
		Success:    true,
	}
}
