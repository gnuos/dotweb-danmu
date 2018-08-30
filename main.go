// danmu project main.go
package main

import (
	stdLog "log"

	"github.com/devfeel/dotweb"
	"github.com/labstack/gommon/log"
)

func main() {
	log.SetLevel(log.DEBUG)

	hub := NewHub()
	go hub.run()

	app := dotweb.New()
	app.SetProductionMode()

	app.HttpServer.GET("/", home)
	app.HttpServer.GET("/chat", chat)
	app.HttpServer.WebSocket("/ws", func(ctx dotweb.Context) error {
		return serveWS(hub, ctx)
	})

	log.Fatal(app.Start())
}

func chat(ctx dotweb.Context) error {
	return ctx.View("chat.html")
}

func home(ctx dotweb.Context) error {
	return ctx.View("index.html")
}

func serveWS(hub *Hub, ctx dotweb.Context) error {
	client := &Client{
		hub:    hub,
		closed: make(chan struct{}),
		send:   make(chan string, 100),
		ws:     ctx.WebSocket(),
	}
	client.hub.register <- client

	stdLog.Println("开始websocket服务...")

	hub.wg.Add(2)

	go client.readLoop()
	go client.writeLoop()

	hub.wg.Wait()

	return nil
}
