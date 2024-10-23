package repository

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/redis/rueidis"
)

var _ RateLimiter = &RateLimitRepository{}

type RateLimitRepository struct {
	client rueidis.Client
}

func NewRateLimiterRepository(redisClient rueidis.Client) *RateLimitRepository {
	return &RateLimitRepository{client: redisClient}
}

func (r *RateLimitRepository) IsAllowed(ctx context.Context, ip string, limit int, window time.Duration) (bool, int, time.Time, error) {
	key := fmt.Sprintf("ratelimit:ip:%s", ip)
	now := time.Now()
	windowStart := now.Add(-window)

	cmds := []rueidis.Completed{
		r.client.B().Zremrangebyscore().
			Key(key).
			Min("0").
			Max(strconv.FormatInt(windowStart.Unix(), 10)).
			Build(),

		r.client.B().Zadd().
			Key(key).
			ScoreMember().
			ScoreMember(float64(now.Unix()), strconv.FormatInt(now.Unix(), 10)).
			Build(),

		r.client.B().Zcard().
			Key(key).
			Build(),

		r.client.B().Expire().
			Key(key).
			Seconds(int64(window.Seconds())).
			Build(),
	}

	resp := r.client.DoMulti(ctx, cmds...)
	for _, result := range resp {
		if err := result.Error(); err != nil {
			return false, 0, time.Time{}, fmt.Errorf("failed to get resp: %w", err)
		}
	}

	count, err := resp[2].ToInt64()
	if err != nil {
		return false, 0, time.Time{}, fmt.Errorf("failed to get count: %w", err)
	}

	remaining := limit - int(count)
	if remaining < 0 {
		remaining = 0
	}

	reset := now.Add(window)
	allowed := count <= int64(limit)

	return allowed, remaining, reset, nil
}
