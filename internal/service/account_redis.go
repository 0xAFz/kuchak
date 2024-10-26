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

func (a *AccountRedisService) SetEmailVerify(ctx context.Context, email, token string) error {
	return a.repo.SaveVerify(ctx, email, token, time.Minute*5)
}

func (a *AccountRedisService) GetByVerifyToken(ctx context.Context, token string) (string, error) {
	return a.repo.ByVerifyToken(ctx, token)
}

func (a *AccountRedisService) GetByVerifyEmail(ctx context.Context, token string) (string, error) {
	return a.repo.ByVerifyToken(ctx, token)
}

func (a *AccountRedisService) SetResetPassword(ctx context.Context, email, token string) error {
	return a.repo.SaveReset(ctx, email, token, time.Minute*5)
}

func (a *AccountRedisService) GetByResetPasswordToken(ctx context.Context, token string) (string, error) {
	return a.repo.ByResetToken(ctx, token)
}

func (a *AccountRedisService) GetByResetPasswordEmail(ctx context.Context, email string) (string, error) {
	return a.repo.ByResetEmail(ctx, email)
}
