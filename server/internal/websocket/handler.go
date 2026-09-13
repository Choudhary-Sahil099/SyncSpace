package websocket

import (
	"fmt"
	"net/http"
	"syncspace/internal/auth"

	"github.com/google/uuid"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func ServeWS(hub *Hub, authService *auth.Service, c *gin.Context) {

	roomID := c.Param("roomId")
	user, err := authService.AuthenticateAccessToken(c.Query("access_token"))
	if err != nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "a valid access token is required"})
		return
	}
	username := user.Username
	if !authService.HasDocumentAccess(user.ID, roomID) {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "you do not have access to this document"})
		return
	}

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
		ID:       uuid.NewString(),
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
