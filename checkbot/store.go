package main

import (
	"sync"

	"github.com/cirias/accessible"
)

type ResultStore struct {
	mutex sync.RWMutex
	items []*accessible.Result
	start int
	len   int
}

func NewResultStore(size int) *ResultStore {
	return &ResultStore{
		items: make([]*accessible.Result, size),
	}
}

func (s *ResultStore) Append(item *accessible.Result) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.items[(s.start+s.len)%len(s.items)] = item
	if s.len < len(s.items) {
		s.len += 1
	} else {
		s.start += 1
		s.start %= len(s.items)
	}
}

func (s *ResultStore) Range(f func(*accessible.Result) bool) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	for i := s.start; i < s.len; i++ {
		item := s.items[(s.start+i)%len(s.items)]
		if !f(item) {
			return
		}
	}
}
