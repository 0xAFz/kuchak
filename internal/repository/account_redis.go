package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/rueidis"
)

var _ AccountRedis = &AccountRedisRepository{}

type AccountRedisRepository struct {
	client rueidis.Client
}

func NewAccountRedisRepository(redisClient rueidis.Client) *AccountRedisRepository {
	return &AccountRedisRepository{client: redisClient}
}

func (a *AccountRedisRepository) SaveVerifyToken(ctx context.Context, email, token string, ttl time.Duration) error {
	key := "verify:" + token
	cmd := a.client.B().Set().Key(key).Value(email).Nx().Px(ttl).Build()

	err := a.client.Do(ctx, cmd).Error()
	if err != nil {
		return fmt.Errorf("failed to set token into redis: %w", err)
	}

	return nil
}

func (a *AccountRedisRepository) ByToken(ctx context.Context, token string) (string, error) {
	key := "verify:" + token
	cmd := a.client.B().Getdel().Key(key).Build()

	result, err := a.client.Do(ctx, cmd).ToString()
	if err != nil {
		if rueidis.IsRedisNil(err) {
			return "", fmt.Errorf("token is not valid or expierd: %w", err)
		}
		return "", fmt.Errorf("failed to get token from redis: %w", err)
	}

	return result, nil
}
