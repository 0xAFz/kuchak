package services

import (
	"kuchak/internal/repository"
)

type URLRedisService struct {
	repo repository.URLRedis
}

func NewURLRedisService(repo repository.URLRedis) *URLRedisService {
	return &URLRedisService{repo: repo}
}
