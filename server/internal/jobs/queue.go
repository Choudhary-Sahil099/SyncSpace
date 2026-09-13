package jobs

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

const QueueName = "syncspace:jobs"

type Job struct {
	Type    string `json:"type"`
	RoomID  string `json:"roomId"`
	Version int64  `json:"version"`
}

type Queue struct{ client *redis.Client }

func New(redisURL string) (*Queue, error) {
	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, fmt.Errorf("parse redis URL: %w", err)
	}
	client := redis.NewClient(opts)
	if err := client.Ping(context.Background()).Err(); err != nil {
		client.Close()
		return nil, err
	}
	return &Queue{client: client}, nil
}
func (q *Queue) Close() error { return q.client.Close() }
func (q *Queue) Enqueue(ctx context.Context, job Job) error {
	payload, err := json.Marshal(job)
	if err != nil {
		return err
	}
	return q.client.RPush(ctx, QueueName, payload).Err()
}
func (q *Queue) Pop(ctx context.Context, timeout time.Duration) (*Job, error) {
	result, err := q.client.BLPop(ctx, timeout, QueueName).Result()
	if err != nil {
		return nil, err
	}
	if len(result) < 2 {
		return nil, nil
	}
	var job Job
	if err := json.Unmarshal([]byte(result[1]), &job); err != nil {
		return nil, err
	}
	return &job, nil
}
