package websocket

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

const (
	writeWait = 10 * time.Second
	pongWait  = 60 * time.Second
	pingPeriod = (pongWait * 9) / 10
	maxMessageSize = 5120
)

type Client struct {
	ID       string
	Username string
	RoomID   string
	Conn     *websocket.Conn
	Send     chan Message

	mu     sync.Mutex
	Closed bool
}

func (c *Client) Close() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.Closed {
		close(c.Send)
		c.Closed = true
	}
}

func (c *Client) ReadMessage(hub *Hub) {

	defer func() {

		fmt.Println("CLIENT DISCONNECTED:", c.Username)

		hub.Unregister <- c

		c.Conn.Close()
	}()

	c.Conn.SetReadLimit(maxMessageSize)

	c.Conn.SetReadDeadline(time.Now().Add(pongWait))

	c.Conn.SetPongHandler(func(string) error {

		c.Conn.SetReadDeadline(time.Now().Add(pongWait))

		return nil
	})

	for {

		_, payload, err := c.Conn.ReadMessage()

		if err != nil {

			fmt.Println("READ ERROR:", err)

			break
		}

		fmt.Println("MESSAGE RECEIVED:", string(payload))

		var message Message

		if err := json.Unmarshal(payload, &message); err != nil {

			fmt.Println("UNMARSHAL ERROR:", err)

			continue
		}

		hub.Broadcast <- message
	}
}

func (c *Client) WriteMessage() {

	ticker := time.NewTicker(pingPeriod)

	defer func() {

		ticker.Stop()

		c.Conn.Close()
	}()

	for {

		select {

		case message, ok := <-c.Send:

			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))

			if !ok {

				fmt.Println("CHANNEL CLOSED")

				c.Conn.WriteMessage(
					websocket.CloseMessage,
					[]byte{},
				)

				return
			}

			fmt.Println("WRITING TO:", c.Username)

			payload, err := json.Marshal(message)

			if err != nil {

				fmt.Println("MARSHAL ERROR:", err)

				continue
			}

			c.mu.Lock()

			err = c.Conn.WriteMessage(
				websocket.TextMessage,
				payload,
			)

			c.mu.Unlock()

			if err != nil {

				fmt.Println("WRITE ERROR:", err)

				return
			}

			fmt.Println("MESSAGE WRITTEN")

		case <-ticker.C:

			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))

			c.mu.Lock()

			err := c.Conn.WriteMessage(
				websocket.PingMessage,
				nil,
			)

			c.mu.Unlock()

			if err != nil {

				fmt.Println("PING ERROR:", err)

				return
			}
		}
	}
}