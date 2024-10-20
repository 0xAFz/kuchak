package service

import (
	"context"
	"kuchak/internal/entity"
	"kuchak/internal/repository"
)

type URLRedisService struct {
	repo repository.URLRedis
}

func NewURLRedisService(repo repository.URLRedis) *URLRedisService {
	return &URLRedisService{repo: repo}
}

func (u *URLRedisService) GetFromCacheByShortURL(ctx context.Context, shortURL string) (entity.URL, error) {
	return u.repo.ByShortURL(ctx, shortURL)
}

func (u *URLRedisService) SetURLToCache(ctx context.Context, url entity.URL) error {
	return u.repo.Save(ctx, url)
}
