package main

import (
	"syncspace/internal/websocket"

	"github.com/gin-gonic/gin"
)

func main() {

	router := gin.Default()

	hub := websocket.NewHub()

	go hub.Run()

	router.GET("/ws/:roomId", func(c *gin.Context) {
		websocket.ServeWS(hub, c)
	})

	router.Run(":8080")
}