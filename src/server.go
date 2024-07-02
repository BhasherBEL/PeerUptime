package main

import (
	"bhasherbel/peeruptime/types"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type Server struct {
	Hosts  *types.Hosts
	Config *types.ServerConfig
}

func (s *Server) Start() {
	s.handlers()
	http.ListenAndServe(s.Config.Ip+":"+s.Config.Port, nil)
}

func (s *Server) handlers() {
	http.HandleFunc("/health", s.healthHandler)
	http.HandleFunc("/api/status", s.statusHandler)
}

func (s *Server) healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("OK"))
}

func (s *Server) statusHandler(w http.ResponseWriter, r *http.Request) {
	request := &types.StatusRequest{}

	json.NewDecoder(r.Body).Decode(request)

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Request-Time", time.Now().UTC().Format(time.RFC3339Nano))

	response := types.Response{
		Status:      "OK",
		Config:      &types.SharedConfig{},
		Discoveries: *hosts.Keys,
	}

	json.NewEncoder(w).Encode(response)

	go func() {
		if request.Discoverable && request.DiscoverableURL != "" && request.DiscoverableURL != s.Config.URL {
			if _, ok := hosts.Get(request.DiscoverableURL); !ok {
				fmt.Println("Discovered new host: " + request.DiscoverableURL)
				hosts.AppendNew(request.DiscoverableURL)
			}
		}
	}()
}
