package main

import botapi "github.com/go-telegram-bot-api/telegram-bot-api"

type State interface {
	Update(msg *botapi.Message) (State, botapi.Chattable)
}

type Server struct {
	chats map[int64]State
}

func NewServer() *Server {
	return &Server{
		chats: make(map[int64]State),
	}
}

func (s *Server) Serve(bot *botapi.BotAPI, initState State, updates <-chan botapi.Update) error {
	// TODO handle updates in parallel
	for u := range updates {
		if u.Message == nil {
			continue
		}

		state, ok := s.chats[u.Message.Chat.ID]
		if !ok {
			state = initState
			s.chats[u.Message.Chat.ID] = state
		}

		nextState, chattable := state.Update(u.Message)
		s.chats[u.Message.Chat.ID] = nextState

		if chattable == nil {
			continue
		}

		_, err := bot.Send(chattable)
		if err != nil {
			return err
		}
	}

	return nil
}
