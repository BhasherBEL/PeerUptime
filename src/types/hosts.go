package types

import (
	"container/heap"
)

type Host struct {
	URL      string
	Priority int
	Checks   *Checks
}

type Hosts struct {
	entries map[string]*Host
	Keys    *[]string
	Size    int
	pq      *PriorityQueue
}

func NewHosts() *Hosts {
	hs := &Hosts{
		entries: make(map[string]*Host),
		Keys:    &[]string{},
		Size:    0,
		pq:      &PriorityQueue{},
	}
	heap.Init(hs.pq)
	return hs
}

func (hs Hosts) Get(url string) (*Host, bool) {
	h, ok := hs.entries[url]
	return h, ok
}

func (hs Hosts) Peek() *Item {
	return hs.pq.Peek().(*Item)
}

func (hs Hosts) UpdatePriority(item *Item, priority int) {
	hs.pq.Update(item, priority)
}

func (hs Hosts) Append(h *Host) {
	hs.entries[h.URL] = h
	*hs.Keys = append(*hs.Keys, h.URL)
	hs.Size++
	hs.pq.Push(&Item{
		Value:    h.URL,
		Priority: scoreCnt,
	})
}

func (hs Hosts) AppendNew(url string) {
	h := &Host{
		URL:      url,
		Priority: scoreCnt,
		Checks: &Checks{
			Entries: make([]*Check, 0),
			Size:    0,
			Score:   0.5,
			Average: 0,
		},
	}
	hs.Append(h)
}
