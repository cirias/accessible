package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"log"
	"net/http"
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

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	httpc := &http.Client{Transport: tr}
	client := &accessible.Client{
		Username:   *username,
		Password:   *password,
		Entrypoint: *entrypoint,
		Httpc:      httpc,
	}

	bot, updates, err := NewBotAndFetch(*token)
	if err != nil {
		log.Fatalln(err)
	}

	initState := NewSubState(client, bot)

	server := NewServer()
	if err := server.Serve(bot, initState, updates); err != nil {
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
