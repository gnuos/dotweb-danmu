package main

import (
	"io"
	stdLog "log"
	"time"

	"github.com/devfeel/dotweb"
	"github.com/labstack/gommon/log"
)

type Client struct {
	hub    *Hub
	closed chan struct{}
	send   chan string
	ws     *dotweb.WebSocket
}

func (c *Client) readLoop() {
	defer func() {
		c.hub.unregister <- c
		c.ws.Conn.Close()
		c.hub.wg.Done()
	}()

	for {
		//判断是否为心跳信息
		message, err := c.ws.ReadMessage()
		if err != nil {
			if err == io.EOF {
				log.Warn("连接被客户端关闭了！")
				c.closed <- struct{}{}
				break
			}

			log.Errorf("error: %v", err)
			break
		}

		switch message {
		case "__PING__":
			if err := c.ws.SendMessage("__PONG__"); err != nil {
				log.Error(err)
				return
			}
		case "__PONG__":
			stdLog.Println(c.ws.Request().RemoteAddr, " is connected!")
		default:
			c.hub.broadcast <- message
		}
	}
}

func (c *Client) writeLoop() {
	var ticker = time.NewTicker(5 * time.Second)
	defer func() {
		ticker.Stop()
		c.ws.Conn.Close()
		c.hub.wg.Done()
	}()

	for {
		select {
		case <-c.closed:
			return
		case message, ok := <-c.send:
			if !ok {
				c.ws.Conn.WriteClose(1000)
				return
			}
			//将消息正式发送给客户端
			err := c.ws.SendMessage(message)
			if err != nil {
				log.Error(err)
			}
		case <-ticker.C: //发送心跳信息
			c.hub.broadcast <- "__PING__"
		default:
			time.Sleep(2 * time.Second)
			//持续向客户端发广播
			c.hub.broadcast <- "浙江温州江南皮革厂倒闭啦 ！！！"
		}
	}
}
