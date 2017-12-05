package main

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/cirias/accessible"
)

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

func (s *Subscriber) Subscriptions() *sync.Map {
	return s.subs
}

func (s *Subscriber) Subscribe(name, url string, d time.Duration, handleAnomaly func(*accessible.Result, error) error) {
	store := NewRecycleStore(1*time.Minute, 100)
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
		if err := client.Poll(ctx, handle, url, d); err != nil {
			log.Println("could not poll: %s", err)
		}
	}()

	sub := &Subscription{
		cancel:   cancel,
		url:      url,
		duration: d,
		history:  store,
	}
	s.subs.Store(name, sub)
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

type Subscription struct {
	cancel   func()
	url      string
	duration time.Duration
	history  *RecycleStore
}

func (s *Subscription) History() []*accessible.Result {
	return s.history.Load()
}

func (s *Subscription) Close() {
	s.cancel()
	s.history.Close()
}
