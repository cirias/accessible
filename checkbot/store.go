package main

import (
	"sync"
	"time"

	"github.com/cirias/accessible"
)

type Store struct {
	// I want to make Store a interface
	// but I can't due to []*Result is not []interface{}
	mutex sync.RWMutex
	items []*accessible.Result
}

func NewStore() *Store {
	return &Store{
		items: make([]*accessible.Result, 0),
	}
}

func (s *Store) Append(item *accessible.Result) {
	s.mutex.Lock()
	s.items = append(s.items, item)
	s.mutex.Unlock()
}

func (s *Store) Load() []*accessible.Result {
	s.mutex.RLock()
	items := s.items[:]
	s.mutex.RUnlock()
	return items
}

func (s *Store) Drop(before int) {
	s.mutex.Lock()
	s.items = s.items[before:]
	s.mutex.Unlock()
}

type RecycleStore struct {
	*Store
	t *time.Ticker
}

func NewRecycleStore(d time.Duration, max int) *RecycleStore {
	s := NewStore()
	t := time.NewTicker(d)

	go func() {
		for range t.C {
			items := s.Load()

			if len(items) <= max {
				continue
			}

			before := len(items) - max
			s.Drop(before)
		}
	}()

	return &RecycleStore{s, t}
}

func (s *RecycleStore) Close() {
	s.t.Stop()
}
