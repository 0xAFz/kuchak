package repository

import (
	"context"
	"kuchak/internal/entity"
)

type Account interface {
	ByID(ctx context.Context, ID int) (entity.User, error)
	ByEmail(ctx context.Context, email string) (entity.User, error)
	Save(ctx context.Context, user entity.User) error
	Delete(ctx context.Context, user entity.User) error
}

type URL interface {
	ByID(ctx context.Context, ID int) (entity.URL, error)
	ByShortURL(ctx context.Context, shortURL string) (entity.URL, error)
	Save(ctx context.Context, url entity.URL) error
	UpdateClickCount(ctx context.Context, shortURL string) error
	Delete(ctx context.Context, url entity.URL) error
}

type AccountRedis interface {
	SaveVerifyEmail(ctx context.Context, email, token string) error
	ByToken(ctx context.Context, token string) entity.VerifyEmail
}

type URLRedis interface {
	ByShortURL(ctx context.Context, shortURL string) (entity.URL, error)
	Save(ctx context.Context, url entity.URL) error
}
