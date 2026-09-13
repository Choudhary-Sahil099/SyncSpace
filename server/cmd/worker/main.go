package main

import (
	"context"
	"log"
	"os"
	"time"

	"syncspace/internal/ai"
	"syncspace/internal/jobs"
	"syncspace/internal/storage"
)

func main() {
	databaseURL := os.Getenv("DATABASE_URL")
	redisURL := os.Getenv("REDIS_URL")
	if databaseURL == "" || redisURL == "" {
		log.Fatal("DATABASE_URL and REDIS_URL are required")
	}

	ctx := context.Background()
	store, err := storage.NewPostgresDocumentStore(ctx, databaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer store.Close()
	queue, err := jobs.New(redisURL)
	if err != nil {
		log.Fatal(err)
	}
	defer queue.Close()

	log.Println("SyncSpace worker started")
	for {
		job, err := queue.Pop(ctx, 30*time.Second)
		if err != nil {
			log.Println("queue error:", err)
			time.Sleep(time.Second)
			continue
		}
		if job == nil {
			continue
		}
		if job.Type != "document.embed" {
			continue
		}
		doc := store.GetDocument(job.RoomID)
		if doc.Content == "" {
			continue
		}
		embedding, model, err := ai.CreateEmbedding(ctx, doc.Content)
		if err != nil {
			log.Println("embedding job failed:", err)
			continue
		}
		if err := store.SaveEmbedding(ctx, job.RoomID, doc.Content, model, embedding); err != nil {
			log.Println("embedding persistence failed:", err)
			continue
		}
		log.Printf("indexed room=%s version=%d model=%s", job.RoomID, doc.Version, model)
	}
}
