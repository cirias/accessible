package main

import (
	"time"

	"github.com/cirias/accessible"
)

type LoadCloser interface {
	Load() []*accessible.Result
	Close()
}

type Subscription struct {
	cancel   func()
	url      string
	duration time.Duration
	history  LoadCloser
}

func (s *Subscription) History() []*accessible.Result {
	return s.history.Load()
}

func (s *Subscription) Close() {
	s.cancel()
	s.history.Close()
}
