package repository

import (
	"context"
	"encoding/json"
	"ecommerce/user-service/internal/domain"
	"time"

	"github.com/go-redis/redis/v8"
)

type sessionRepo struct{ Redis *redis.Client }

func NewRedisSessionRepo(client *redis.Client) domain.SessionRepository {
	return &sessionRepo{Redis: client}
}

func (r *sessionRepo) CreateSession(ctx context.Context, token string, data *domain.SessionData) error {
	// Convert the struct to a JSON string
	bytes, err := json.Marshal(data)
	if err != nil {
		return err
	}
	
	// Save the JSON string in Redis
	return r.Redis.Set(ctx, token, string(bytes), 24*time.Hour).Err()
}

func (r *sessionRepo) GetSessionData(ctx context.Context, token string) (*domain.SessionData, error) {
	// Fetch the JSON string
	val, err := r.Redis.Get(ctx, token).Result()
	if err != nil {
		return nil, err // Returns redis.Nil if expired
	}

	// Convert the JSON string back into the struct
	var data domain.SessionData
	if err := json.Unmarshal([]byte(val), &data); err != nil {
		return nil, err
	}

	return &data, nil
}