package service

import (
	"context"
	"kuchak/internal/repository"
	"time"
)

type AccountRedisService struct {
	repo repository.AccountRedis
}

func NewAccountRedisService(repo repository.AccountRedis) *AccountRedisService {
	return &AccountRedisService{repo: repo}
}

func (a *AccountRedisService) SetEmailVerifyToken(ctx context.Context, email, token string) error {
	return a.repo.SaveVerifyToken(ctx, email, token, time.Minute*5)
}

func (a *AccountRedisService) GetByToken(ctx context.Context, token string) (string, error) {
	return a.repo.ByToken(ctx, token)
}
