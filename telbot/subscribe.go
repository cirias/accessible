package main

import (
	"context"
	"strings"
	"sync"
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

func (s *Subscription) Close() {
	s.cancel()
	s.history.Close()
}

type SubHandler struct {
	client *accessible.Client
	mux    sync.RWMutex
	subs   map[int64]map[string]*Subscription
}

func (h *SubHandler) add(chatId int64, s *Subscription) {
	h.mux.Lock()
	defer h.mux.Unlock()

	subs, ok := h.subs[chatId]
	if !ok {
		subs = make(map[string]*Subscription)
		h.subs[chatId] = subs
	}
	subs[s.url] = s
}

func (h *SubHandler) ServeCommand(bot *botapi.BotAPI, m *botapi.Message) {
	args := parseArguments(m.CommandArguments())
	if len(args) != 2 {
		// TODO invalid command arguments
		return
	}

	url := args[0]

	duration, err := time.ParseDuration(args[1])
	if err != nil {
		// TODO invalid command arguments
		return
	}

	results := make(chan *accessible.Result)
	defer close(results)

	// TODO maybe should put these in NewSubscription
	ctx, cancel := context.WithCancel(context.Background())
	s := &Subscription{
		cancel:   cancel,
		url:      url,
		duration: duration,
		history:  NewRecycleStore(1*time.Minute, 100),
	}

	h.add(m.Chat.ID, s)

	go func() {
		for r := range results {
			s.history.Append(r)

			if r.Success() {
				continue
			}

			// TODO send alert
		}
	}()

	h.client.Poll(ctx, results, url, duration)
	// TODO success subscribed
}

func (h *SubHandler) HandleList(bot *botapi.BotAPI, m *botapi.Message) {
	h.mux.RLock()
	defer h.mux.RUnlock()

	// for k, v := range h.subs[m.Chat.ID] {
	// TODO output like
	// `
	// <url> <duration>
	// <url> <duration>
	// ...
	// `
	// }
}

func (h *SubHandler) HandleRemove(bot *botapi.BotAPI, m *botapi.Message) {
	h.mux.RLock()
	defer h.mux.RUnlock()

	args := parseArguments(m.CommandArguments())
	if len(args) != 1 {
		// TODO invalid command arguments
		return
	}

	name := args[0]

	h.subs[m.Chat.ID][name].Close()
	delete(h.subs[m.Chat.ID], args[0])
	// TODO return success
}

func parseArguments(rawArgs string) []string {
	args := strings.Split(rawArgs, " ")
	trimedArgs := make([]string, 0)
	for _, arg := range args {
		trimed := strings.Trim(arg, " ")
		if trimed == "" {
			continue
		}

		trimedArgs = append(trimedArgs, trimed)
	}

	return trimedArgs
}
