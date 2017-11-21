package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/cirias/accessible"
)

var (
	username = flag.String("username", "", "optional simple authentication username")
	password = flag.String("password", "", "optional simple authentication password")
	laddr    = flag.String("laddr", ":7654", "address that http serve listen to")
	cert     = flag.String("cert", "cert.pem", "TLS cert path")
	key      = flag.String("key", "key.pem", "TLS key path")
)

func main() {
	flag.Parse()

	handler := accessible.BasicAuthHandler(*username, *password, accessible.CheckHandler)
	http.Handle("/check", handler)
	if err := http.ListenAndServeTLS(*laddr, *cert, *key, nil); err != nil {
		log.Fatalln(err)
	}
}
