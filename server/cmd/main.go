package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"syncspace/internal/ai"
	"syncspace/internal/auth"
	"syncspace/internal/jobs"
	"syncspace/internal/metrics"
	"syncspace/internal/relay"
	"syncspace/internal/search"
	"syncspace/internal/storage"
	"syncspace/internal/websocket"
	"syncspace/internal/workspace"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func main() {
	router := gin.Default()
	router.Use(corsMiddleware())

	databasePath := envOr("SYNCSPACE_DB_PATH", "syncspace.db")
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
	workspace.RegisterRoutes(router, authService)

	// Use PostgreSQL when configured; otherwise retain the zero-dependency memory store.
	var store storage.DocumentStore = storage.NewDocumentStore()
	var closeStore func()
	if databaseURL := os.Getenv("DATABASE_URL"); databaseURL != "" {
		postgresStore, err := storage.NewPostgresDocumentStore(context.Background(), databaseURL)
		if err != nil {
			log.Fatal("initialize postgres document store: ", err)
		}
		store = postgresStore
		closeStore = postgresStore.Close
		log.Println("document persistence: PostgreSQL")
	} else {
		log.Println("document persistence: in-memory (set DATABASE_URL for PostgreSQL)")
	}
	if closeStore != nil {
		defer closeStore()
	}

	var redisRelay *relay.RedisRelay
	var jobQueue *jobs.Queue
	if redisURL := os.Getenv("REDIS_URL"); redisURL != "" {
		redisRelay, err = relay.New(redisURL, uuid.NewString())
		if err != nil {
			log.Fatal("initialize redis relay: ", err)
		}
		defer redisRelay.Close()
		log.Println("realtime fanout: Redis Pub/Sub")
		jobQueue, err = jobs.New(redisURL)
		if err != nil {
			log.Fatal("initialize job queue: ", err)
		}
		defer jobQueue.Close()
		log.Println("background jobs: Redis queue")
	} else {
		log.Println("realtime fanout: process-local (set REDIS_URL for Redis Pub/Sub)")
	}

	hub := websocket.NewHubWithStore(store, redisRelay, jobQueue)
	go hub.Run()
	router.GET("/ws/:roomId", func(c *gin.Context) { websocket.ServeWS(hub, authService, c) })
	search.RegisterRoutes(router, store, authService)
	ai.RegisterRoutes(router, store, authService)

	m := &metrics.Metrics{}
	router.GET("/api/health", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"status": "ok"}) })
	router.GET("/metrics", gin.WrapH(http.HandlerFunc(m.Handler)))

	log.Println("SyncSpace API listening on :8080")
	if err := router.Run(":8080"); err != nil {
		log.Fatal(err)
	}
}

func envOr(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		if origin == "" {
			origin = "*"
		}
		c.Header("Access-Control-Allow-Origin", origin)
		c.Header("Vary", "Origin")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	}
}
