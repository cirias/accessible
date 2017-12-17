package main

import (
	"fmt"
	"sync"

	"github.com/cirias/accessible"
	"github.com/cirias/accessible/telbot"
)

type Handler interface {
	Handle(*telbot.Bot, *telbot.Update)
}

type Server struct {
	h Handler
}

func NewServer(h Handler) *Server {
	return &Server{
		h: h,
	}
}

func (s *Server) Serve(bot *telbot.Bot) error {
	params := &telbot.GetUpdatesParams{
		Offset:  0,
		Limit:   100,
		Timeout: 10,
	}
	for {
		updates, err := bot.GetUpdates(params)
		if err != nil {
			return fmt.Errorf("could not get updates: %s", err)
		}

		for _, u := range updates {
			go s.h.Handle(bot, u)
		}

		if len(updates) > 0 {
			params.Offset = updates[len(updates)-1].Id + 1
		}
	}
}

type ChatHandler struct {
	chats  *sync.Map
	client *accessible.Client
}

func NewChatHandler(client *accessible.Client) *ChatHandler {
	return &ChatHandler{
		chats:  &sync.Map{},
		client: client,
	}
}

func (h *ChatHandler) Handle(bot *telbot.Bot, u *telbot.Update) {
	chatId := u.Message.Chat.Id
	v, ok := h.chats.Load(chatId)
	if !ok {
		v = NewChat(chatId, h.client)
		h.chats.Store(chatId, v)
	}

	go v.(Handler).Handle(bot, u)
}
