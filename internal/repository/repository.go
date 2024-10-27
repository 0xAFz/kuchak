package repository

import (
	"context"
	"kuchak/internal/entity"
	"time"
)

type Account interface {
	ByID(ctx context.Context, ID int) (entity.User, error)
	ByEmail(ctx context.Context, email string) (entity.User, error)
	Save(ctx context.Context, user entity.User) error
	Delete(ctx context.Context, user entity.User) error
	UpdateEmail(ctx context.Context, user entity.User) error
	UpdatePassword(ctx context.Context, user entity.User) error
	UpdateVerifyEmail(ctx context.Context, user entity.User) error
}

type URL interface {
	ByID(ctx context.Context, ID int) (entity.URL, error)
	ByShortURL(ctx context.Context, shortURL string) (entity.URL, error)
	ByUserID(ctx context.Context, userID int) ([]entity.URL, error)
	Save(ctx context.Context, url entity.URL) error
	UpdateClickCount(ctx context.Context, shortURL string) error
	Delete(ctx context.Context, url entity.URL) error
}

type AccountRedis interface {
	ByVerifyEmail(ctx context.Context, token string) (string, error)
	ByVerifyToken(ctx context.Context, email string) (string, error)
	SaveVerify(ctx context.Context, email, token string, ttl time.Duration) error

	ByResetEmail(ctx context.Context, token string) (string, error)
	ByResetToken(ctx context.Context, email string) (string, error)
	SaveReset(ctx context.Context, email, token string, ttl time.Duration) error
}

type URLRedis interface {
	ByShortURL(ctx context.Context, shortURL string) (entity.URL, error)
	Save(ctx context.Context, url entity.URL) error
}

type RateLimiter interface {
	IsAllowed(ctx context.Context, ip string, limit int, window time.Duration) (bool, int, time.Time, error)
}
