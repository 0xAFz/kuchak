package services

import (
	"kuchak/internal/repository"
)

type URLPostgresService struct {
	repo repository.URL
}

func NewURLPostgresService(repo repository.URL) *URLPostgresService {
	return &URLPostgresService{repo: repo}
}
