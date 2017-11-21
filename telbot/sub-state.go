package main

import (
	"fmt"
	"time"

	botapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

type SubState struct {
	subConfig struct {
		Name     string
		URL      string
		Duration time.Duration
	}
	newSub func(chatId int64, name, url string, d time.Duration) *Subscription
	subs   map[string]*Subscription
	update func(SubState, *botapi.Message) (SubState, botapi.Chattable)
}

func NewSubState(newSub func(chatId int64, name, url string, d time.Duration) *Subscription) SubState {
	return SubState{
		newSub: newSub,
		subs:   make(map[string]*Subscription),
	}
}

func (s SubState) Update(msg *botapi.Message) (State, botapi.Chattable) {
	return s.update(s, msg)
}

func UpdateIdle(s SubState, msg *botapi.Message) (SubState, botapi.Chattable) {
	switch msg.Command() {
	case "sub":
		return s, botapi.NewMessage(msg.Chat.ID, "Name?")
		/*
		 * case "rm":
		 *   return nil, nil
		 * case "ls":
		 *   return nil, nil
		 */
	}

	return s, nil
}

func UpdateSubNew(s SubState, msg *botapi.Message) (SubState, botapi.Chattable) {
	if msg.Command() == "cancel" {
		s.update = UpdateIdle
		return s, botapi.NewMessage(msg.Chat.ID, "Canceled")
	}

	/*
	 * if msg.IsCommand() {
	 *   return s, botapi.NewMessage(msg.Chat.ID, "Invalid name")
	 * }
	 */

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

	sub := s.newSub(msg.Chat.ID, name, url, duration)
	s.subs[s.subConfig.Name] = sub

	s.update = UpdateIdle
	return s, botapi.NewMessage(msg.Chat.ID, "Created")
}