package relay

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

type RedisRelay struct {
	client *redis.Client
	pubsub *redis.PubSub
	Origin string
}

func New(redisURL, origin string) (*RedisRelay, error) {
	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, fmt.Errorf("parse redis URL: %w", err)
	}
	client := redis.NewClient(opts)
	ctx := context.Background()
	if err := client.Ping(ctx).Err(); err != nil {
		client.Close()
		return nil, fmt.Errorf("redis ping: %w", err)
	}
	pubsub := client.Subscribe(ctx, "syncspace:rooms")
	return &RedisRelay{client: client, pubsub: pubsub, Origin: origin}, nil
}

func (r *RedisRelay) Close() error { r.pubsub.Close(); return r.client.Close() }
func (r *RedisRelay) Publish(payload []byte) error {
	return r.client.Publish(context.Background(), "syncspace:rooms", payload).Err()
}
func (r *RedisRelay) Messages() <-chan *redis.Message { return r.pubsub.Channel() }
