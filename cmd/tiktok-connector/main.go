package main

import (
	"context"
	"flag"
	"log"
	"net/http"

	"tiktok-connector/connector"
)

func main() {
	addr := flag.String("addr", "127.0.0.1:8787", "HTTP address for the local connector")
	remoteBase := flag.String("remote-base", connector.DefaultRemoteRelayBase, "Remote relay base URL")
	remoteTopic := flag.String("remote-topic", connector.DefaultRemoteTopic, "Remote relay topic; set empty to disable")
	flag.Parse()

	server := connector.NewServer()
	if *remoteTopic != "" {
		subscribeURL := connector.RemoteSubscribeURL(*remoteBase, *remoteTopic)
		connector.StartRemoteRelay(context.Background(), server.Hub(), subscribeURL)
		log.Printf("Remote connector relay: %s", subscribeURL)
	}
	log.Printf("TikTok connector: http://%s", *addr)
	log.Fatal(http.ListenAndServe(*addr, server.Routes()))
}
