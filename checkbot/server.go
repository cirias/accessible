package main

import (
	"fmt"
	"sync"

	"github.com/cirias/accessible"
	"github.com/cirias/accessible/telbot"
)

type HandleFunc func(telbot.SendMessageFunc, *telbot.Update)

type Handler interface {
	Handle(telbot.SendMessageFunc, *telbot.Update)
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
	params := &telbot.UpdatesParams{}
	for {
		// TODO offset
		updates, err := bot.GetUpdates(params)
		if err != nil {
			return fmt.Errorf("could not get updates: %s", err)
		}

		for _, u := range updates {
			go s.h.Handle(bot.SendMessage, u)
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

func (h *ChatHandler) Handle(s telbot.SendMessageFunc, u *telbot.Update) {
	chatId := u.Message.Chat.Id
	v, ok := h.chats.Load(chatId)
	if !ok {
		v = NewChat(h.client)
		h.chats.Store(chatId, v)
	}

	go v.(Handler).Handle(s, u)
}
