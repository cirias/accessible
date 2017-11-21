package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/cirias/accessible"
	botapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

type Subscription struct {
	cancel   func()
	url      string
	duration time.Duration
	history  *RecycleStore
}

func NewSubscription(client *accessible.Client, bot *botapi.BotAPI, url string, d time.Duration, chatId int64) *Subscription {
	ctx, cancel := context.WithCancel(context.Background())
	results := make(chan *accessible.Result)
	go func() {
		defer close(results)
		if err := client.Poll(ctx, results, url, d); err != nil {
			log.Println("could not poll: %s", err)
		}
	}()

	store := NewRecycleStore(1*time.Minute, 100)
	go func() {
		for r := range results {
			store.Append(r)

			if r.Success() {
				continue
			}

			msg := botapi.NewMessage(chatId, fmt.Sprint(*r))
			if _, err := bot.Send(msg); err != nil {
				log.Println("could not send: %s", err)
			}
		}
	}()

	return &Subscription{
		cancel:   cancel,
		url:      url,
		duration: d,
		history:  store,
	}
}

func (s *Subscription) History() []*accessible.Result {
	return s.history.Load()
}

func (s *Subscription) Close() {
	s.cancel()
	s.history.Close()
}
