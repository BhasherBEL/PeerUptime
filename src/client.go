package main

import (
	"bhasherbel/peeruptime/types"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type Client struct {
	Hosts  *types.Hosts
	Config *types.ClientConfig
}

func (c *Client) Start() {
	for {
		c.check()
		time.Sleep(time.Duration(c.Config.WaitingTime) * time.Millisecond)
	}
}

func (c *Client) check() {
	host, ok := c.Hosts.Peek()
	if !ok {
		fmt.Println("No hosts to check")
		return
	}

	check := c.checkHost(host)

	if !check.Success && (host.Checks.Last() == nil || *host.Checks.Last()) {
		host.UnseenTime = &check.Time
	}

	if host.Checks.Last() == nil || check.Success != *host.Checks.Last() {
		if check.Success {
			if host.UnseenTime != nil {
				offineTime := time.Now().UTC().Sub(*host.UnseenTime)
				fmt.Println("Host " + host.URL + " is back online after " + offineTime.String())
			} else {
				fmt.Println("Host " + host.URL + " is online for the first time")
			}
		} else {
			fmt.Println("Host " + host.URL + " is offline")
		}
	}

	host.Checks.Append(check, float64(c.Config.MemoryFactor))
}

func (c *Client) checkHost(host *types.Host) types.Check {
	before := time.Now().UTC()

	request := types.StatusRequest{
		Discovery:          true,
		Discoverable:       true,
		DiscoverableURL:    c.Config.DiscoverableURL,
		DiscoverableConfig: &types.SharedConfig{},
	}

	jsonData, err := json.Marshal(request)
	if err != nil {
		return types.Check{Time: before, Success: false}
	}

	resp, err := http.Post(host.URL+"/api/status", "application/json", bytes.NewBuffer(jsonData))
	after := time.Now().UTC()
	if err != nil {
		return types.Check{Time: before, Success: false}
	}
	defer resp.Body.Close()

	pingTime, err := time.Parse(time.RFC3339, resp.Header.Get("X-Request-Time"))
	if err != nil {
		return types.Check{Time: before, Success: false}
	}

	response := &types.Response{}

	json.NewDecoder(resp.Body).Decode(response)

	for _, discovery := range response.Discoveries {
		if discovery == c.Config.DiscoverableURL {
			continue
		}
		if _, ok := hosts.Get(discovery); !ok {
			fmt.Println("Discovered new host: " + discovery)
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
