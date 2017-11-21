package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/cirias/accessible"
	botapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

var (
	username   = flag.String("username", "", "username to access check server")
	password   = flag.String("password", "", "password to access check server")
	entrypoint = flag.String("entrypoint", "", "entrypoint of check server")
	token      = flag.String("token", "", "telegram bot token")
)

func main() {
	flag.Parse()

	bot, updates, err := NewBotAndFetch(*token)
	if err != nil {
		log.Fatalln(err)
	}

	client := accessible.NewClient(*username, *password, *entrypoint)

	newSub := func(chatId int64, name, url string, d time.Duration) *Subscription {
		ctx, cancel := context.WithCancel(context.Background())
		results := make(chan *accessible.Result)
		go func() {
			defer close(results)
			client.Poll(ctx, results, url, d)
		}()

		store := NewRecycleStore(1*time.Minute, 100)
		go func() {
			for r := range results {
				store.Append(r)

				if r.Success() {
					continue
				}

				msg := botapi.NewMessage(chatId, fmt.Sprint(r))
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

	newState := func() State {
		return NewSubState(newSub)
	}

	if err := Serve(bot, newState, updates); err != nil {
		log.Fatalln(err)
	}
}

func NewBotAndFetch(token string) (*botapi.BotAPI, <-chan botapi.Update, error) {
	bot, err := botapi.NewBotAPI(token)
	if err != nil {
		return nil, nil, fmt.Errorf("could not new bot: %s", err)
	}
	bot.Debug = true

	u := botapi.NewUpdate(0)
	u.Timeout = 60
	updates, err := bot.GetUpdatesChan(u)
	if err != nil {
		return nil, nil, fmt.Errorf("could not get updates channel: %s", err)
	}

	// Optional: wait for updates and clear them if you don't want to handle
	// a large backlog of old messages
	time.Sleep(time.Millisecond * 500)
	updates.Clear()

	return bot, updates, nil
}
