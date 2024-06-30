package main

import (
	"container/heap"
	"fmt"
	"math"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"strconv"
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
}

type Check struct {
	Time       time.Time
	PingDelay  float64
	PongDelay  float64
	LocalDelay float64
	Success    bool
}

type Checks struct {
	Entries []*Check
	Size    int
	Score   float64
	Average float64
}

func (cs Checks) Append(c Check) {
	cs.Entries = append(cs.Entries, &c)
	cs.Average = (cs.Average*float64(cs.Size) + boolToFloat(c.Success)) / float64(cs.Size+1)

	reverseFactor := 1. / float64(config.MemoryScoreFactor)
	cs.Score = cs.Average*(1.-reverseFactor) + boolToFloat(c.Success)*reverseFactor
}

func (cs Checks) AmortizedScore() float64 {
	score := float64(scoreCnt)
	score += float64(config.VariableScoreFactor) * math.Abs(cs.Score-0.5)
	score += float64(config.VariableScoreFactor) * rand.NormFloat64()
	return score
}

type Host struct {
	URL      string
	Priority int
	Checks   *Checks
}

// Helpers

func intOrDefault(value string, defaultValue int) int {
	if value == "" {
		return defaultValue
	}

	intValue, err := strconv.Atoi(value)
	if err != nil {
		return defaultValue
	} else {
		return intValue
	}
}

func stringOrDefault(value string, defaultValue string) string {
	if value == "" {
		return defaultValue
	}
	return value
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

func boolToFloat(b bool) float64 {
	if b {
		return 1.
	}
	return 0.
}

// Global variables

var config = Config{
	DiscoveryLimit:      intOrDefault(os.Getenv("PEER_DISCOVERY_LIMIT"), 5),
	Ip:                  stringOrDefault(os.Getenv("PEER_IP"), "0.0.0.0"),
	Port:                stringOrDefault(os.Getenv("PEER_PORT"), "8080"),
	URL:                 stringOrDefault(os.Getenv("PEER_URL"), "http://localhost:8080"),
	DiscoveryURL:        stringOrDefault(os.Getenv("PEER_DISCOVERY_URL"), "http://localhost:8081"),
	MemoryScoreFactor:   intOrDefault(os.Getenv("PEER_MEMORY_SCORE_FACTOR"), 10),
	VariableScoreFactor: intOrDefault(os.Getenv("PEER_VARIABLE_SCORE_FACTOR"), 1000),
	WaitingTime:         intOrDefault(os.Getenv("PEER_WAITING_TIME"), 1000),
}

var hosts = map[string]Host{
	config.DiscoveryURL: {
		URL:      config.DiscoveryURL,
		Priority: 0,
		Checks:   &Checks{Entries: make([]*Check, 0), Size: 0, Score: 0.5, Average: 0},
	},
}

var pq = make(PriorityQueue, len(hosts))

var scoreCnt = 0

// Main

func main() {
	i := 0
	for url := range hosts {
		pq[i] = &Item{
			value:    url,
			priority: scoreCnt,
			index:    i,
		}
		i++
	}
	heap.Init(&pq)

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK"))
	})

	http.HandleFunc("/api/status", status)

	go http.ListenAndServe(config.Ip+":"+config.Port, nil)
	fmt.Println("Server started on " + config.Ip + ":" + config.Port)

	go loopCheck()
	fmt.Println("Check loop started")

	sigint := make(chan os.Signal, 1)
	signal.Notify(sigint, os.Interrupt)

	<-sigint

	fmt.Println("Received SIGINT signal. Shuting down...")
	os.Exit(0)
}

func status(w http.ResponseWriter, r *http.Request) {
	//params := r.URL.Query()
	//discovery := params.Get("discovery") == "true"
	//discoveryLimit := intOrDefault(params.Get("discoveryLimit"), config.DiscoveryLimit)

	w.Header().Set("X-Request-Time", time.Now().UTC().Format(time.RFC3339Nano))

	w.Write([]byte("OK"))
}

func loopCheck() {
	for {
		check()
		time.Sleep(time.Duration(config.WaitingTime) * time.Millisecond)
	}
}

func check() {
	scoreCnt++

	item := pq.Peek().(*Item)
	url := item.value
	host := hosts[url]

	fmt.Print("Checking " + url + " ... ")

	check := checkHost(host)
	host.Checks.Append(check)

	if check.Success {
		fmt.Printf("success (%vms, %vms, %vms)\n", check.PingDelay, check.PongDelay, check.LocalDelay)
	} else {
		fmt.Println("Failed")
	}

	newPriority := int(host.Checks.AmortizedScore())

	pq.update(item, newPriority)
}

func checkHost(host Host) Check {
	before := time.Now().UTC()

	resp, err := http.Get(host.URL + "/api/status")
	after := time.Now().UTC()
	if err != nil {
		return Check{Time: before, Success: false}
	}
	defer resp.Body.Close()

	pingTime, err := time.Parse(time.RFC3339, resp.Header.Get("X-Request-Time"))
	if err != nil {
		return Check{Time: before, Success: false}
	}

	return Check{
		Time:       before,
		PingDelay:  float64(pingTime.Sub(before).Microseconds()) / 1000.,
		PongDelay:  float64(after.Sub(pingTime).Microseconds()) / 1000.,
		LocalDelay: float64(after.Sub(before).Microseconds()) / 1000.,
		Success:    true,
	}
}
