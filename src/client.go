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
	item := c.Hosts.Peek()
	url := item.Value
	host, _ := hosts.Get(url)

	fmt.Printf("Checking %v ... ", url)

	check := c.checkHost(host)
	host.Checks.Append(check, float64(c.Config.MemoryFactor))

	if check.Success {
		fmt.Printf("success (%vms, %vms, %vms)\n", check.PingDelay, check.PongDelay, check.LocalDelay)
	} else {
		fmt.Println("Failed")
	}

	newPriority := int(host.Checks.AmortizedScore())

	hosts.UpdatePriority(item, newPriority)
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

	response := &types.Response{}

	json.NewDecoder(resp.Body).Decode(response)

	for _, discovery := range response.Discoveries {
		if discovery == c.Config.DiscoverableURL {
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
