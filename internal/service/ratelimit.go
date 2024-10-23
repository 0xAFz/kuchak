package service

import (
	"context"
	"kuchak/internal/repository"
	"time"
)

type RateLimitService struct {
	repo repository.RateLimiter
}

func NewRateLimitService(repo repository.RateLimiter) *RateLimitService {
	return &RateLimitService{repo: repo}
}

func (r *RateLimitService) IsAllowed(ctx context.Context, ip string, limit int, window time.Duration) (bool, int, time.Time, error) {
	return r.repo.IsAllowed(ctx, ip, limit, window)
}
