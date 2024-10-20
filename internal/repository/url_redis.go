package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"kuchak/internal/entity"
	"time"

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
	jsonData, err := json.Marshal(url)
	if err != nil {
		return fmt.Errorf("failed to serialize URL: %w", err)
	}

	key := "url:" + url.ShortURL

	cmd := u.client.B().Set().Key(key).Value(string(jsonData)).Ex(time.Second * 3600).Build()

	err = u.client.Do(ctx, cmd).Error()
	if err != nil {
		return fmt.Errorf("failed to save url into redis: %w", err)
	}

	return nil
}

func (u *URLRedisRepository) ByShortURL(ctx context.Context, shortURL string) (entity.URL, error) {
	key := fmt.Sprintf("url:%s", shortURL)
	cmd := u.client.B().Get().Key(key).Build()

	jsonData, err := u.client.Do(ctx, cmd).ToString()
	if err != nil {
		if rueidis.IsRedisNil(err) {
			return entity.URL{}, fmt.Errorf("url not found")
		}
		return entity.URL{}, fmt.Errorf("failed to get url from redis: %w", err)
	}

	var url entity.URL
	err = json.Unmarshal([]byte(jsonData), &url)
	if err != nil {
		return entity.URL{}, fmt.Errorf("failed to deserialize url: %w", err)
	}

	return url, nil
}
