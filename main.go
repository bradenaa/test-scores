package main

import (
	"fmt"
	"net/http"

	"github.com/r3labs/sse"
)

func handleEvents(events chan *sse.Event) {
	// 	// Got some data!
	for {
		msg := <-events
		fmt.Printf("%v", string(msg.Data))
	}
}

func hello(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("hello!"))
}

func main() {
	// handling event stream in a new channel
	events := make(chan *sse.Event)
	client := sse.NewClient("http://live-test-scores.herokuapp.com/scores")
	client.SubscribeChan("messages", events)
	go handleEvents(events)

	// establishing a server
	http.HandleFunc("/", hello)
	http.ListenAndServe(":8080", nil)
}
