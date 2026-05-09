// internal/infrastructure/cache/redis/session_store.go
package redis

import (
	"context"
	"time"

	"auth-service/internal/domain/port/session"
)

type SessionStore struct {
	client *Client
}

func NewSessionStore(c *Client) session.SessionStore {
	return &SessionStore{client: c}
}

func (s *SessionStore) Set(ctx context.Context, key, value string, exp time.Duration) error {
	return s.client.client.Set(ctx, key, value, exp).Err()
}

func (s *SessionStore) Get(ctx context.Context, key string) (string, error) {
	return s.client.client.Get(ctx, key).Result()
}

func (s *SessionStore) Delete(ctx context.Context, key string) error {
	return s.client.client.Del(ctx, key).Err()
}
