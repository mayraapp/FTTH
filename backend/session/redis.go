// Package session administra sesiones compartidas para las réplicas FTTH.
package session

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"ftth/backend/domain"
	"github.com/redis/go-redis/v9"
)

const keyPrefix = "ftth:session:"

// Store permite intercambiar la sesión local por una sesión global en Redis.
type Store interface {
	Create(context.Context, string, domain.User, time.Duration) error
	Get(context.Context, string) (domain.User, bool, error)
	Delete(context.Context, string) error
	Close() error
}

// RedisStore conserva sólo un identificador aleatorio en la cookie; los datos
// del usuario permanecen del lado del servidor y expiran junto con la sesión.
type RedisStore struct {
	client *redis.Client
}

func NewRedisStore(ctx context.Context, address, password string) (*RedisStore, error) {
	client := redis.NewClient(&redis.Options{Addr: address, Password: password, DB: 0})
	if err := client.Ping(ctx).Err(); err != nil {
		_ = client.Close()
		return nil, fmt.Errorf("no se pudo conectar a Redis: %w", err)
	}
	return &RedisStore{client: client}, nil
}

func (s *RedisStore) Create(ctx context.Context, id string, account domain.User, ttl time.Duration) error {
	payload, err := json.Marshal(account)
	if err != nil {
		return fmt.Errorf("no se pudo serializar la sesión: %w", err)
	}
	return s.client.Set(ctx, keyPrefix+id, payload, ttl).Err()
}

func (s *RedisStore) Get(ctx context.Context, id string) (domain.User, bool, error) {
	payload, err := s.client.Get(ctx, keyPrefix+id).Bytes()
	if err == redis.Nil {
		return domain.User{}, false, nil
	}
	if err != nil {
		return domain.User{}, false, err
	}
	var account domain.User
	if err := json.Unmarshal(payload, &account); err != nil {
		return domain.User{}, false, fmt.Errorf("sesión de Redis inválida: %w", err)
	}
	return account, true, nil
}

func (s *RedisStore) Delete(ctx context.Context, id string) error {
	return s.client.Del(ctx, keyPrefix+id).Err()
}

func (s *RedisStore) Close() error {
	return s.client.Close()
}
