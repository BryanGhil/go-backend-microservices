package repository

import (
	"context"
	"ecommerce/user-service/internal/domain"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
)

type sessionRepo struct{ Redis *redis.Client }

func NewRedisSessionRepo(client *redis.Client) domain.SessionRepository {
	return &sessionRepo{Redis: client}
}

func (r *sessionRepo) StoreRefreshToken(ctx context.Context, token string, data *domain.SessionData) error {
	tracer := otel.Tracer("redis-repository")
	ctx, span := tracer.Start(ctx, "Redis.StoreRefreshToken")
	defer span.End()

	bytes, err := json.Marshal(data)
	if err != nil {
		span.RecordError(err)
		return err
	}

	expiration := 7 * 24 * time.Hour

	// 1. Save the actual session data
	err = r.Redis.Set(ctx, token, string(bytes), expiration).Err()
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to save refresh token")
		return err
	}

	// 2. Add this token to the User's list of active sessions
	userSetKey := fmt.Sprintf("user_sessions:%d", data.UserID)
	err = r.Redis.SAdd(ctx, userSetKey, token).Err()
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to add token to user set")
		return err
	}

	return nil
}

func (r *sessionRepo) GetSessionData(ctx context.Context, token string) (*domain.SessionData, error) {
	tracer := otel.Tracer("redis-repository")
	ctx, span := tracer.Start(ctx, "Redis.GetSessionData")
	defer span.End()

	val, err := r.Redis.Get(ctx, token).Result()
	if err != nil {
		if err != redis.Nil { // Don't record cache misses as hard errors
			span.RecordError(err)
			span.SetStatus(codes.Error, "redis get failed")
		}
		return nil, err
	}

	var data domain.SessionData
	if err := json.Unmarshal([]byte(val), &data); err != nil {
		span.RecordError(err)
		return nil, err
	}

	return &data, nil
}

func (r *sessionRepo) DeleteRefreshToken(ctx context.Context, token string) error {
	tracer := otel.Tracer("redis-repository")
	ctx, span := tracer.Start(ctx, "Redis.DeleteRefreshToken")
	defer span.End()

	data, err := r.GetSessionData(ctx, token)
	if err != nil {
		return err // Token already gone
	}

	userSetKey := fmt.Sprintf("user_sessions:%d", data.UserID)

	// FIX: Use .Result() to get the actual number of deleted keys!
	// This is an ATOMIC operation.
	deletedCount, err := r.Redis.Del(ctx, token).Result()
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to delete token key")
		return err
	}

	// RACE CONDITION SHIELD: 
	// If deletedCount is 0, another request (React Strict Mode) 
	// beat us to it by 1 millisecond. Abort immediately!
	if deletedCount == 0 {
		errConcurrent := errors.New("token already consumed by concurrent request")
		span.RecordError(errConcurrent)
		return errConcurrent
	}

	// Only the request that actually deleted the token gets to run this cleanup
	if err := r.Redis.SRem(ctx, userSetKey, token).Err(); err != nil {
		span.RecordError(err)
	}

	return nil
}

func (r *sessionRepo) GetUserSessions(ctx context.Context, userID int64) ([]*domain.SessionData, error) {
	tracer := otel.Tracer("redis-repository")
	ctx, span := tracer.Start(ctx, "Redis.GetUserSessions")
	defer span.End()

	userSetKey := fmt.Sprintf("user_sessions:%d", userID)

	tokens, err := r.Redis.SMembers(ctx, userSetKey).Result()
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to get session members")
		return nil, err
	}

	var sessions []*domain.SessionData
	for _, token := range tokens {
		data, err := r.GetSessionData(ctx, token)
		if err == nil {
			sessions = append(sessions, data)
		} else {
			// Clean up expired tokens
			r.Redis.SRem(ctx, userSetKey, token)
		}
	}

	return sessions, nil
}

// --- OTP Methods ---
func (r *sessionRepo) StoreOTP(ctx context.Context, email string, otp string) error {
	tracer := otel.Tracer("redis-repository")
	ctx, span := tracer.Start(ctx, "Redis.StoreOTP")
	defer span.End()

	err := r.Redis.Set(ctx, "otp:"+email, otp, 5*time.Minute).Err()
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to store otp")
		return err
	}
	return nil
}

func (r *sessionRepo) GetOTP(ctx context.Context, email string) (string, error) {
	tracer := otel.Tracer("redis-repository")
	ctx, span := tracer.Start(ctx, "Redis.GetOTP")
	defer span.End()

	otp, err := r.Redis.Get(ctx, "otp:"+email).Result()
	if err != nil && err != redis.Nil {
		span.RecordError(err)
	}
	return otp, err
}

func (r *sessionRepo) DeleteOTP(ctx context.Context, email string) error {
	tracer := otel.Tracer("redis-repository")
	ctx, span := tracer.Start(ctx, "Redis.DeleteOTP")
	defer span.End()

	err := r.Redis.Del(ctx, "otp:"+email).Err()
	if err != nil {
		span.RecordError(err)
	}
	return err
}

// --- Revocation Methods ---
func (r *sessionRepo) RevokeAllSessions(ctx context.Context, userID int64) error {
	tracer := otel.Tracer("redis-repository")
	ctx, span := tracer.Start(ctx, "Redis.RevokeAllSessions")
	defer span.End()

	userSetKey := fmt.Sprintf("user_sessions:%d", userID)

	tokens, err := r.Redis.SMembers(ctx, userSetKey).Result()
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to get user sessions for revocation")
		return err
	}

	for _, token := range tokens {
		r.Redis.Del(ctx, token)
	}

	err = r.Redis.Del(ctx, userSetKey).Err()
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to delete user set key")
		return err
	}

	return nil
}

func (r *sessionRepo) RevokeSession(ctx context.Context, userID int64, sessionID string) error {
	tracer := otel.Tracer("redis-repository")
	ctx, span := tracer.Start(ctx, "Redis.RevokeSession")
	defer span.End()

	userSetKey := fmt.Sprintf("user_sessions:%d", userID)
	tokens, err := r.Redis.SMembers(ctx, userSetKey).Result()
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to fetch members")
		return err
	}

	for _, token := range tokens {
		data, err := r.GetSessionData(ctx, token)
		if err == nil && data.SessionID == sessionID {
			r.Redis.Del(ctx, token)
			r.Redis.SRem(ctx, userSetKey, token)
			return nil
		}
	}
	
	errNotFound := errors.New("session not found")
	span.RecordError(errNotFound)
	return errNotFound
}