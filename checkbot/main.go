package main

import (
	"crypto/tls"
	"flag"
	"log"
	"net/http"

	"github.com/cirias/accessible"
	"github.com/cirias/accessible/telbot"
)

var (
	username   = flag.String("username", "", "username to access check server")
	password   = flag.String("password", "", "password to access check server")
	entrypoint = flag.String("entrypoint", "", "entrypoint of check server")

	token = flag.String("token", "", "telegram bot token")
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
	ch := NewChatHandler(client)
	server := NewServer(ch)

	bot := telbot.NewBot(*token)
	if err := server.Serve(bot); err != nil {
		log.Fatalln(err)
	}
}
