package main

import botapi "github.com/go-telegram-bot-api/telegram-bot-api"

type State interface {
	Update(msg *botapi.Message) (State, botapi.Chattable)
}

var chats map[int64]State = make(map[int64]State)

func Serve(bot *botapi.BotAPI, newState func() State, updates <-chan botapi.Update) error {
	// TODO handle updates in parallel
	for u := range updates {
		if u.Message == nil {
			continue
		}

		state, ok := chats[u.Message.Chat.ID]
		if !ok {
			state = newState()
			chats[u.Message.Chat.ID] = state
		}

		nextState, chattable := state.Update(u.Message)
		chats[u.Message.Chat.ID] = nextState

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
