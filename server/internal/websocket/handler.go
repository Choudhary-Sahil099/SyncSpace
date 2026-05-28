package websocket

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func ServeWS(hub *Hub, c *gin.Context) {

	roomID := c.Param("roomId")

	username := c.Query("username")

	conn, err := upgrader.Upgrade(
		c.Writer,
		c.Request,
		nil,
	)

	if err != nil {
		return
	}

	client := &Client{
		ID:       conn.RemoteAddr().String(),
		Username: username,
		RoomID:   roomID,
		Conn:     conn,

		Send: make(chan Message, 256),
	}

	hub.Register <- client

	go client.WriteMessage()

	go client.ReadMessage(hub)
}