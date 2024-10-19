package repository

import (
	"context"
	"kuchak/internal/entity"

	"github.com/redis/rueidis"
)

var _ AccountRedis = &AccountRedisRepository{}

type AccountRedisRepository struct {
	client rueidis.Client
}

func NewAccountRedisRepository(redisClient rueidis.Client) *AccountRedisRepository {
	return &AccountRedisRepository{client: redisClient}
}

func (a *AccountRedisRepository) SaveVerifyEmail(ctx context.Context, email, token string) error {
	panic("...")
}

func (a *AccountRedisRepository) ByToken(ctx context.Context, token string) entity.VerifyEmail {
	panic("...")
}
