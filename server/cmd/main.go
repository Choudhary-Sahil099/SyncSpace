package main

import (
	"log"
	"os"
	"syncspace/internal/auth"
	"syncspace/internal/websocket"

	"github.com/gin-gonic/gin"
)

func main() {

	router := gin.Default()
	router.Use(corsMiddleware())

	databasePath := os.Getenv("SYNCSPACE_DB_PATH")
	if databasePath == "" {
		databasePath = "syncspace.db"
	}
	jwtSecret := os.Getenv("SYNCSPACE_JWT_SECRET")
	if jwtSecret == "" {
		var err error
		jwtSecret, err = auth.NewDevelopmentSecret()
		if err != nil {
			log.Fatal("create development JWT secret: ", err)
		}
		log.Println("SYNCSPACE_JWT_SECRET is not set; sessions will end when the server restarts")
	}

	authService, err := auth.NewService(databasePath, jwtSecret)
	if err != nil {
		log.Fatal("initialize authentication: ", err)
	}
	defer authService.Close()
	auth.RegisterRoutes(router, authService)

	hub := websocket.NewHub()

	go hub.Run()

	router.GET("/ws/:roomId", func(c *gin.Context) {

		websocket.ServeWS(hub, c)
	})

	router.Run(":8080")
}

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", c.Request.Header.Get("Origin"))
		c.Header("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	}
}
