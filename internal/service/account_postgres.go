package services

import (
	"kuchak/internal/repository"
)

type AccountPostgresService struct {
	repo repository.Account
}

func NewAccountPostgresService(repo repository.Account) *AccountPostgresService {
	return &AccountPostgresService{repo: repo}
}
