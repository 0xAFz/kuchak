package repository

import (
	"context"
	"kuchak/internal/entity"

	"github.com/redis/rueidis"
)

var _ URLRedis = &URLRedisRepository{}

type URLRedisRepository struct {
	client rueidis.Client
}

func NewURLRedisRepository(redisClient rueidis.Client) *URLRedisRepository {
	return &URLRedisRepository{client: redisClient}
}

func (u *URLRedisRepository) Save(ctx context.Context, url entity.URL) error {
	panic("...")
}

func (u *URLRedisRepository) ByShortURL(ctx context.Context, shortURL string) (entity.URL, error) {
	panic("...")
}
