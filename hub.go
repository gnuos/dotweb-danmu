package main

import (
	"sync"
)

type Hub struct {
	wg sync.WaitGroup

	// Registered clients.
	clients map[*Client]bool

	// Inbound messages from the clients.
	broadcast chan string

	// Register requests from the clients.
	register chan *Client

	// Unregister requests from clients.
	unregister chan *Client
}

func NewHub() *Hub {
	return &Hub{
		broadcast:  make(chan string),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[*Client]bool),
	}
}

func (h *Hub) run() {
	for {
		select {
		case client := <-h.register:
			//注册客户端
			h.clients[client] = true
		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				//移除客户端
				delete(h.clients, client)
			}
		case message := <-h.broadcast:
			for client := range h.clients {
				//向每个客户端发送消息
				client.send <- message
			}
		}
	}
}
