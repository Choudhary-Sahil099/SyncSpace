package websocket

import (
	"github.com/gorilla/websocket"
	"encoding/json"
	"fmt"
)

type Client struct {
	ID       string
	Username string
	RoomID   string
	Conn     *websocket.Conn
	Send     chan Message
}
func (c *Client) ReadMessage(hub *Hub) {

	defer func() {
		hub.Unregister <- c
		c.Conn.Close()
	}()

	for {

		_, payload, err := c.Conn.ReadMessage()

		if err != nil {
			break
		}

		var message Message

		if err := json.Unmarshal(payload, &message); err != nil {
			continue
		}

		hub.Broadcast <- message
	}
}
func (c *Client) WriteMessage() {

	defer c.Conn.Close()

	for {

		message, ok := <-c.Send

		if !ok {
			fmt.Println("CHANNEL CLOSED")
			return
		}

		fmt.Println("WRITING TO:", c.Username)

		payload, err := json.Marshal(message)

		if err != nil {
			fmt.Println("MARSHAL ERROR:", err)
			continue
		}

		err = c.Conn.WriteMessage(websocket.TextMessage, payload)

		if err != nil {
			fmt.Println("WRITE ERROR:", err)
			return
		}

		fmt.Println("MESSAGE WRITTEN")
	}
}