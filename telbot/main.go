package main

import (
	"flag"
	"log"
)

func main() {
	token := flag.String("token", "", "telegram bot token")
	flag.Parse()

	subh := &SubHandler{}

	mux := NewServeMux()
	mux.Handle("sub", subh)
	if err := mux.NewBotAndServe(*token); err != nil {
		log.Fatalln(err)
	}
}
