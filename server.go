package main

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// a chat server manages transports and peers and all that jazz
type chatServer struct {
	whomst string
	chat   *chat

	http     http.Server
	upgrader websocket.Upgrader
}

func newChatServer(whomst, addr string) *chatServer {
	s := &chatServer{
		whomst: whomst,
		chat:   &chat{},
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/chat", s.handlePeer)
	s.http = http.Server{
		Addr:    addr,
		Handler: mux,
	}

	// use the defaults
	s.upgrader = websocket.Upgrader{}

	return s
}

func (c *chatServer) propagateChanges(wg sync.WaitGroup, conn *websocket.Conn) {
	defer wg.Done()

	ticker := time.NewTicker(20 * time.Millisecond)
	defer ticker.Stop()

	var lastRead clock
	for _ = range ticker.C {
		changes := c.chat.since(lastRead)
		lastRead = lastRead.update(changes.lastMessageAt())

		if err := conn.WriteJSON(&changes); err != nil {
			fmt.Println("error: somebody is outtttta here: ", err)
			return
		}
	}
}

func (c *chatServer) mergeIncomingChanges(wg sync.WaitGroup, conn *websocket.Conn) {
	defer wg.Done()

	for {
		var changes chat
		if err := conn.ReadJSON(&changes); err != nil {
			fmt.Println("error: somebody is outtttta here: ", err)
			return
		}
		c.chat.merge(&changes)
	}
}

func (c *chatServer) Dial(peer string) {
	conn, _, err := websocket.DefaultDialer.Dial(peer, nil)
	if err != nil {
		panic(err)
	}

	// FIXME: since most errors have been connection closes, this relies on both
	// the sender and receiver encountering errors at the same time. it's real
	// bad.
	wg := sync.WaitGroup{}
	wg.Add(2)
	go c.propagateChanges(wg, conn)
	go c.mergeIncomingChanges(wg, conn)
	wg.Wait()
}

func (c *chatServer) ListenAndServe() error {
	return c.http.ListenAndServe()
}

func (c *chatServer) handlePeer(w http.ResponseWriter, r *http.Request) {
	conn, err := c.upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Fprintf(w, "websocket error : %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// FIXME: since most errors have been connection closes, this relies on both
	// the sender and receiver encountering errors at the same time. it's real
	// bad.
	wg := sync.WaitGroup{}
	wg.Add(2)
	go c.propagateChanges(wg, conn)
	go c.mergeIncomingChanges(wg, conn)
	wg.Wait()
}
