package websocket

import (
	"net/http"
	"fmt"
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

    fmt.Println("WS CONNECT REQUEST:", username, roomID)

    conn, err := upgrader.Upgrade(
        c.Writer,
        c.Request,
        nil,
    )

    if err != nil {
        fmt.Println("WS UPGRADE ERROR:", err)
        return
    }

    fmt.Println(
        "WS UPGRADED:",
        username,
        conn.RemoteAddr().String(),
    )

    client := &Client{
        ID:       conn.RemoteAddr().String(),
        Username: username,
        RoomID:   roomID,
        Conn:     conn,
        Send:     make(chan Message, 256),
    }

    fmt.Println("REGISTERING CLIENT:", client.Username)

    hub.Register <- client

    fmt.Println("CLIENT REGISTERED:", client.Username)

    go client.WriteMessage()
    go client.ReadMessage(hub)
}