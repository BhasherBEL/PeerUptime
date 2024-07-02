package types

import (
	"time"
)

type Host struct {
	URL        string
	Priority   int
	Checks     *Checks
	UnseenTime *time.Time
}

type Hosts struct {
	entries map[string]*Host
	Keys    *[]string
	Size    int
	current int
}

func NewHosts() *Hosts {
	hs := &Hosts{
		entries: make(map[string]*Host),
		Keys:    &[]string{},
		Size:    0,
		current: 0,
	}
	return hs
}

func (hs *Hosts) Get(url string) (*Host, bool) {
	h, ok := hs.entries[url]
	return h, ok
}

func (hs *Hosts) Peek() (*Host, bool) {
	if hs.Size == 0 {
		return nil, false
	}

	h, o := hs.Get((*hs.Keys)[hs.current])

	hs.current++
	if hs.current >= hs.Size {
		hs.current = 0
	}

	return h, o
}

func (hs *Hosts) Append(h *Host) {
	hs.entries[h.URL] = h
	*hs.Keys = append(*hs.Keys, h.URL)
	hs.Size++
}

func (hs *Hosts) AppendNew(url string) {
	h := &Host{
		URL:      url,
		Priority: scoreCnt,
		Checks: &Checks{
			Entries: make([]*Check, 0),
			Size:    0,
			Score:   0.5,
			Average: 0,
		},
		UnseenTime: nil,
	}
	hs.Append(h)
}
