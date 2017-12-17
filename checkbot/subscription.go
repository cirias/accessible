package main

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/cirias/accessible"
)

type SubcribeParams struct {
	Name     string
	URL      string
	Duration time.Duration
}

func (p SubcribeParams) String() string {
	return fmt.Sprintf("%s %s %v", p.Name, p.URL, p.Duration)
}

type Subscription struct {
	cancel  func()
	params  SubcribeParams
	history *ResultStore
}

func (s *Subscription) String() string {
	return fmt.Sprint(s.params)
}

func (s *Subscription) History() *ResultStore {
	return s.history
}

func (s *Subscription) Close() {
	s.cancel()
}

type Subscriber struct {
	client *accessible.Client
	subs   *sync.Map
}

func NewSubscriber(client *accessible.Client) *Subscriber {
	return &Subscriber{
		client: client,
		subs:   &sync.Map{},
	}
}

func (s *Subscriber) Subscribe(p SubcribeParams, handleAnomaly func(*accessible.Result, error) error) {
	store := NewResultStore(100)
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		handle := func(r *accessible.Result, err error) error {
			if err != nil {
				return handleAnomaly(r, err)
			}

			store.Append(r)

			if r.Success() {
				return nil
			}

			return handleAnomaly(r, nil)
		}
		if err := s.client.Poll(ctx, handle, p.URL, p.Duration); err != nil {
			log.Println("could not poll: %s", err)
		}
	}()

	sub := &Subscription{
		cancel:  cancel,
		params:  p,
		history: store,
	}
	s.subs.Store(p.Name, sub)
}

func (s *Subscriber) Unsubscribe(name string) error {
	sub, ok := s.subs.Load(name)
	if !ok {
		return fmt.Errorf("could not found")
	}

	s.subs.Delete(name)
	sub.(*Subscription).Close()

	return nil
}
