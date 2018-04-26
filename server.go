package main

import (
	"fmt"
	"net/http"
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

func (c *chatServer) propagateChanges(conn *websocket.Conn) {
	ticker := time.NewTicker(20 * time.Millisecond)
	defer ticker.Stop()

	var lastRead clock
	for _ = range ticker.C {
		changes := c.chat.since(lastRead)
		lastRead = lastRead.update(changes.lastMessageAt())

		if err := conn.WriteJSON(&changes); err != nil {
			panic(err)
		}
	}
}

func (c *chatServer) mergeIncomingChanges(conn *websocket.Conn) {
	for {
		var changes chat
		if err := conn.ReadJSON(&changes); err != nil {
			panic(err)
		}
		c.chat.merge(&changes)
	}
}

func (c *chatServer) Dial(peer string) {
	conn, _, err := websocket.DefaultDialer.Dial(peer, nil)
	if err != nil {
		panic(err)
	}

	go c.propagateChanges(conn)
	c.mergeIncomingChanges(conn)
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

	go c.propagateChanges(conn)
	c.mergeIncomingChanges(conn)
}
