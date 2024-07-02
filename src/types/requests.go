package types

type Response struct {
	Status      string
	Config      *SharedConfig
	Discoveries []string
}

type SharedConfig struct {
}

type StatusRequest struct {
	Discovery          bool
	Discoverable       bool
	DiscoverableURL    string
	DiscoverableConfig *SharedConfig
}
