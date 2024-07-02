package types

import (
	"bhasherbel/peeruptime/utils"
	"os"
)

type ServerConfig struct {
	Enable bool
	Ip     string
	Port   string
	URL    string
}

type ClientConfig struct {
	Enable          bool
	DiscoveryURL    string
	DiscoverableURL string
	MemoryFactor    int
	WaitingTime     int // ms
}

type Config struct {
	Server *ServerConfig
	Client *ClientConfig
}

func NewConfig() *Config {
	return &Config{
		Server: &ServerConfig{
			Enable: utils.BoolOrDefault(os.Getenv("PEER_SERVER"), true),
			Ip:     utils.StringOrDefault(os.Getenv("PEER_IP"), "0.0.0.0"),
			Port:   utils.StringOrDefault(os.Getenv("PEER_PORT"), "8080"),
			URL:    utils.StringOrDefault(os.Getenv("PEER_URL"), "http://127.0.0.1:8080"),
		},
		Client: &ClientConfig{
			Enable:          utils.BoolOrDefault(os.Getenv("PEER_CLIENT"), true),
			DiscoveryURL:    utils.StringOrDefault(os.Getenv("PEER_DISCOVERY_URL"), "http://localhost:8081"),
			DiscoverableURL: utils.StringOrDefault(os.Getenv("PEER_DISCOVERABLE_URL"), utils.StringOrDefault(os.Getenv("PEER_URL"), "http://127.0.0.1:8080")),
			MemoryFactor:    utils.IntOrDefault(os.Getenv("PEER_MEMORY_SCORE_FACTOR"), 10),
			WaitingTime:     utils.IntOrDefault(os.Getenv("PEER_WAITING_TIME"), 1000),
		},
	}
}
