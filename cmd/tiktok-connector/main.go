package main

import (
	"flag"
	"log"
	"net/http"

	"tiktok-connector/connector"
)

func main() {
	addr := flag.String("addr", "127.0.0.1:8787", "HTTP address for the local connector")
	flag.Parse()

	server := connector.NewServer()
	log.Printf("TikTok connector: http://%s", *addr)
	log.Fatal(http.ListenAndServe(*addr, server.Routes()))
}
