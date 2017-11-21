package main

import (
	"fmt"
	"time"

	"github.com/cirias/accessible"
	botapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

type SubState struct {
	client    *accessible.Client
	bot       *botapi.BotAPI
	subConfig struct {
		Name     string
		URL      string
		Duration time.Duration
	}
	subs   map[string]*Subscription
	update func(SubState, *botapi.Message) (SubState, botapi.Chattable)
}

func NewSubState(client *accessible.Client, bot *botapi.BotAPI) SubState {
	return SubState{
		client: client,
		bot:    bot,
		subs:   make(map[string]*Subscription),
		update: UpdateIdle,
	}
}

func (s SubState) Update(msg *botapi.Message) (State, botapi.Chattable) {
	return s.update(s, msg)
}

func UpdateIdle(s SubState, msg *botapi.Message) (SubState, botapi.Chattable) {
	switch msg.Command() {
	case "sub":
		s.update = UpdateSubNew
		return s, botapi.NewMessage(msg.Chat.ID, "Name?")
	case "ls":
		return s, botapi.NewMessage(msg.Chat.ID, fmt.Sprint(s.subs))
	case "tail":
		arg := msg.CommandArguments()
		return s, botapi.NewMessage(msg.Chat.ID, fmt.Sprint(s.subs[arg].History()))
	case "rm":
		arg := msg.CommandArguments()
		if sub, ok := s.subs[arg]; ok {
			delete(s.subs, arg)
			sub.Close()
		}
		return s, botapi.NewMessage(msg.Chat.ID, "Removed")
	}

	return s, nil
}

func UpdateSubNew(s SubState, msg *botapi.Message) (SubState, botapi.Chattable) {
	if msg.Command() == "cancel" {
		s.update = UpdateIdle
		return s, botapi.NewMessage(msg.Chat.ID, "Canceled")
	}

	s.subConfig.Name = msg.Text // TODO validation
	s.update = UpdateSubName
	return s, botapi.NewMessage(msg.Chat.ID, "URL?")
}

func UpdateSubName(s SubState, msg *botapi.Message) (SubState, botapi.Chattable) {
	if msg.Command() == "cancel" {
		s.update = UpdateIdle
		return s, botapi.NewMessage(msg.Chat.ID, "Canceled")
	}

	s.subConfig.URL = msg.Text // TODO validation
	s.update = UpdateSubURL
	return s, botapi.NewMessage(msg.Chat.ID, "Duration?")
}

func UpdateSubURL(s SubState, msg *botapi.Message) (SubState, botapi.Chattable) {
	if msg.Command() == "cancel" {
		s.update = UpdateIdle
		return s, botapi.NewMessage(msg.Chat.ID, "Canceled")
	}

	duration, err := time.ParseDuration(msg.Text)
	if err != nil {
		return s, botapi.NewMessage(msg.Chat.ID, fmt.Sprintf("could not parse duration: %s", err))
	}

	s.subConfig.Duration = duration
	name := s.subConfig.Name
	url := s.subConfig.URL

	sub := NewSubscription(s.client, s.bot, url, duration, msg.Chat.ID)
	s.subs[name] = sub

	s.update = UpdateIdle
	return s, botapi.NewMessage(msg.Chat.ID, "Created")
}
