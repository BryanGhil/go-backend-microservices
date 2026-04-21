package repository

import (
	"context"
	"ecommerce/user-service/internal/domain"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
)

type sessionRepo struct { Redis *redis.Client }

func NewRedisSessionRepo(client *redis.Client) domain.SessionRepository {
	return &sessionRepo{Redis: client}
}

// Save token with a 24-hour expiration
func (r *sessionRepo) CreateSession(ctx context.Context, token string, userID int64) error {
	return r.Redis.Set(ctx, token, userID, 24*time.Hour).Err()
}

func (r *sessionRepo) GetUserIDByToken(ctx context.Context, token string) (int64, error) {
	val, err := r.Redis.Get(ctx, token).Result()
	if err != nil {
		return 0, err // Returns redis.Nil if token doesn't exist or expired
	}
	return strconv.ParseInt(val, 10, 64)
}