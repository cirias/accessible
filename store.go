package main

import "sync"

type Store struct {
	mutex sync.RWMutex
	items []interface{}
}

func NewStore() *Store {
	return &Store{
		items: make([]interface{}, 0),
	}
}

func (s *Store) Append(item interface{}) {
	s.mutex.Lock()
	s.items = append(s.items, item)
	s.mutex.Unlock()
}

func (s *Store) Load() []interface{} {
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
