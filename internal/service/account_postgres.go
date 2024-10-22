package service

import (
	"context"
	"kuchak/internal/entity"
	"kuchak/internal/repository"
)

type AccountPostgresService struct {
	repo repository.Account
}

func NewAccountPostgresService(repo repository.Account) *AccountPostgresService {
	return &AccountPostgresService{repo: repo}
}

func (a *AccountPostgresService) GetUserByID(ctx context.Context, ID int) (entity.User, error) {
	return a.repo.ByID(ctx, ID)
}

func (a *AccountPostgresService) GetUserByEmail(ctx context.Context, email string) (entity.User, error) {
	return a.repo.ByEmail(ctx, email)
}

func (a *AccountPostgresService) CreateUser(ctx context.Context, user entity.User) error {
	return a.repo.Save(ctx, user)
}

func (a *AccountPostgresService) DeleteUser(ctx context.Context, user entity.User) error {
	return a.repo.Delete(ctx, user)
}

func (a *AccountPostgresService) UpdateUserEmail(ctx context.Context, user entity.User) error {
	return a.repo.UpdateEmail(ctx, user)
}

func (a *AccountPostgresService) UpdateUserPassword(ctx context.Context, user entity.User) error {
	return a.repo.UpdatePassword(ctx, user)
}

func (a *AccountPostgresService) UpdateUserVerifyEmail(ctx context.Context, user entity.User) error {
	return a.repo.UpdateVerifyEmail(ctx, user)
}
