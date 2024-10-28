package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"kuchak/internal/entity"
	"time"

	"github.com/redis/rueidis"
	"github.com/rs/zerolog/log"
)

var _ URLRedis = &URLRedisRepository{}

type URLRedisRepository struct {
	client rueidis.Client
}

func NewURLRedisRepository(redisClient rueidis.Client) *URLRedisRepository {
	return &URLRedisRepository{client: redisClient}
}

func (u *URLRedisRepository) Save(ctx context.Context, url entity.URL) error {
	jsonData, err := json.Marshal(url)
	if err != nil {
		log.Err(err).Msg("failed to serialize url")
		return fmt.Errorf("failed to serialize url: %w", err)
	}

	key := "url:" + url.ShortURL

	cmd := u.client.B().Set().Key(key).Value(string(jsonData)).Ex(time.Second * 3600).Build()

	err = u.client.Do(ctx, cmd).Error()
	if err != nil {
		log.Err(err).Interface("url", url).Msg("failed to set url in redis")
		return fmt.Errorf("failed to set url in redis: %w", err)
	}

	return nil
}

func (u *URLRedisRepository) ByShortURL(ctx context.Context, shortURL string) (entity.URL, error) {
	key := fmt.Sprintf("url:%s", shortURL)
	cmd := u.client.B().Get().Key(key).Build()

	jsonData, err := u.client.Do(ctx, cmd).ToString()
	if err != nil {
		log.Err(err).Str("short_url", shortURL).Msg("failed to fetch url from redis")
		if rueidis.IsRedisNil(err) {
			return entity.URL{}, fmt.Errorf("url not found")
		}
		return entity.URL{}, fmt.Errorf("failed to fetch url from redis: %w", err)
	}

	var url entity.URL
	err = json.Unmarshal([]byte(jsonData), &url)
	if err != nil {
		log.Err(err).Str("short_url", shortURL).Msg("failed to deserialize url")
		return entity.URL{}, fmt.Errorf("failed to deserialize url: %w", err)
	}

	return url, nil
}
