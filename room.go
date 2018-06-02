package main

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

type room struct {
	clients map[*client]struct{}
	join    chan *client
	leave   chan *client
	forward chan []byte
}

func newRoom() *room {
	return &room{
		clients: make(map[*client]struct{}),
		join:    make(chan *client),
		leave:   make(chan *client),
		forward: make(chan []byte),
	}
}

func (r *room) run() {
	for {
		select {
		case c := <-r.join:
			r.clients[c] = struct{}{}
		case c := <-r.leave:
			delete(r.clients, c)
			close(c.send)
		case msg := <-r.forward:
			for c := range r.clients {
				select {
				case c.send <- msg:
				default:
					delete(r.clients, c)
					close(c.send)
				}
			}
		}
	}
}

var upgrader = &websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func (r *room) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	s, err := upgrader.Upgrade(w, req, nil)
	if err != nil {
		log.Fatal(err)
		return
	}
	c := &client{
		socket: s,
		room:   r,
		send:   make(chan []byte, 256),
	}
	r.join <- c
	defer func() { r.leave <- c }()
	go c.write()
	c.read()
}
