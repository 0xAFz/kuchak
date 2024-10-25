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

func (a *AccountRedisService) GetByEmailVerifyToken(ctx context.Context, token string) (string, error) {
	return a.repo.ByVerifyToken(ctx, token)
}

func (a *AccountRedisService) SetResetPassword(ctx context.Context, email, token string) error {
	return a.repo.SaveResetPassword(ctx, email, token, time.Minute*5)
}

func (a *AccountRedisService) GetByResetPasswordToken(ctx context.Context, token string) (string, error) {
	return a.repo.ByResetPasswordToken(ctx, token)
}

func (a *AccountRedisService) GetByResetPasswordEmail(ctx context.Context, email string) (string, error) {
	return a.repo.ByResetPasswordEmail(ctx, email)
}
