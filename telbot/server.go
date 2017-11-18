package main

import (
	"fmt"
	"log"
	"time"

	botapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

type ServeMux struct {
	m map[string]Handler
}

type Handler interface {
	ServeCommand(*botapi.BotAPI, *botapi.Message)
}

func NewServeMux() *ServeMux {
	return &ServeMux{
		m: make(map[string]Handler),
	}
}

func (mux *ServeMux) Handle(command string, h Handler) {
	mux.m[command] = h
}

func (mux *ServeMux) Serve(bot *botapi.BotAPI, updates <-chan botapi.Update) error {
	for u := range updates {
		if u.Message == nil {
			continue
		}

		if !u.Message.IsCommand() {
			continue
		}

		cmd := u.Message.Command()
		h, ok := mux.m[cmd]
		if !ok {
			continue
		}

		go h.ServeCommand(bot, u.Message)
	}

	return nil
}

func (mux *ServeMux) NewBotAndServe(token string) error {
	bot, err := botapi.NewBotAPI(token)
	if err != nil {
		log.Fatalf("could not new bot: %s", err)
	}
	bot.Debug = true

	u := botapi.NewUpdate(0)
	u.Timeout = 60
	updates, err := bot.GetUpdatesChan(u)
	if err != nil {
		return fmt.Errorf("could not get updates channel: %s", err)
	}

	// Optional: wait for updates and clear them if you don't want to handle
	// a large backlog of old messages
	time.Sleep(time.Millisecond * 500)
	updates.Clear()

	return mux.Serve(bot, updates)
}
