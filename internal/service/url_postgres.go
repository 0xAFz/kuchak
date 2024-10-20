package service

import (
	"context"
	"kuchak/internal/entity"
	"kuchak/internal/repository"
)

type URLPostgresService struct {
	repo repository.URL
}

func NewURLPostgresService(repo repository.URL) *URLPostgresService {
	return &URLPostgresService{repo: repo}
}

func (u *URLPostgresService) GetURLByID(ctx context.Context, ID int) (entity.URL, error) {
	return u.repo.ByID(ctx, ID)
}

func (u *URLPostgresService) GetURLByShortURL(ctx context.Context, shortURL string) (entity.URL, error) {
	return u.repo.ByShortURL(ctx, shortURL)
}

func (u *URLPostgresService) CreateURL(ctx context.Context, url entity.URL) error {
	return u.repo.Save(ctx, url)
}

func (u *URLPostgresService) DeleteURL(ctx context.Context, url entity.URL) error {
	return u.repo.Delete(ctx, url)
}

func (u *URLPostgresService) UpdateURLClickCount(ctx context.Context, shortURL string) error {
	return u.repo.UpdateClickCount(ctx, shortURL)
}
