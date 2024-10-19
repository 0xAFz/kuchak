package services

import (
	"kuchak/internal/repository"
)

type AccountRedisService struct {
	repo repository.AccountRedis
}

func NewAccountRedisService(repo repository.AccountRedis) *AccountRedisService {
	return &AccountRedisService{repo: repo}
}
