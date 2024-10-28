package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/rueidis"
	"github.com/rs/zerolog/log"
)

var _ AccountRedis = &AccountRedisRepository{}

type AccountRedisRepository struct {
	client rueidis.Client
}

func NewAccountRedisRepository(redisClient rueidis.Client) *AccountRedisRepository {
	return &AccountRedisRepository{client: redisClient}
}

func (a *AccountRedisRepository) SaveVerify(ctx context.Context, email, token string, ttl time.Duration) error {
	keyEmail := "verify:email:" + email
	keyToken := "verify:token:" + token

	if err := a.client.Do(ctx, a.client.B().Set().Key(keyEmail).Value(email).Nx().Px(ttl).Build()).Error(); err != nil {
		log.Err(err).Str("email", email).Msg("failed to set verify email in redis")
		return fmt.Errorf("failed to set verify email in redis: %w", err)
	}

	if err := a.client.Do(ctx, a.client.B().Set().Key(keyToken).Value(email).Nx().Px(ttl).Build()).Error(); err != nil {
		a.client.Do(ctx, a.client.B().Del().Key(keyEmail).Build())
		log.Err(err).Str("token", token).Msg("failed to set verify token in redis")
		return fmt.Errorf("failed to set verify token in redis: %w", err)
	}

	return nil
}

func (a *AccountRedisRepository) ByVerifyToken(ctx context.Context, token string) (string, error) {
	key := "verify:token:" + token
	cmd := a.client.B().Getdel().Key(key).Build()

	result, err := a.client.Do(ctx, cmd).ToString()
	if err != nil {
		log.Err(err).Msg("failed to fetch verify token from redis")
		if rueidis.IsRedisNil(err) {
			return "", fmt.Errorf("verify token is not valid or expierd: %w", err)
		}
		return "", fmt.Errorf("failed to fetch token from redis: %w", err)
	}

	return result, nil
}

func (a *AccountRedisRepository) ByVerifyEmail(ctx context.Context, email string) (string, error) {
	key := "verify:email:" + email
	cmd := a.client.B().Get().Key(key).Build()

	result, err := a.client.Do(ctx, cmd).ToString()
	if err != nil {
		log.Err(err).Msg("failed to fetch verify email from redis")
		if rueidis.IsRedisNil(err) {
			return "", fmt.Errorf("verify email is not valid or expierd: %w", err)
		}
		return "", fmt.Errorf("failed to fetch verify email from redis: %w", err)
	}

	return result, nil
}

func (a *AccountRedisRepository) SaveReset(ctx context.Context, email, token string, ttl time.Duration) error {
	keyEmail := "reset_password:email:" + email
	keyToken := "reset_password:token:" + token

	if err := a.client.Do(ctx, a.client.B().Set().Key(keyEmail).Value(email).Nx().Px(ttl).Build()).Error(); err != nil {
		log.Err(err).Str("email", email).Msg("failed to set reset password email in redis")
		return fmt.Errorf("failed to set reset password email in redis: %w", err)
	}

	if err := a.client.Do(ctx, a.client.B().Set().Key(keyToken).Value(email).Nx().Px(ttl).Build()).Error(); err != nil {
		log.Err(err).Msg("failed to set reset password token in redis")
		a.client.Do(ctx, a.client.B().Del().Key(keyEmail).Build())
		return fmt.Errorf("failed to set reset password token in redis: %w", err)
	}

	return nil
}

func (a *AccountRedisRepository) ByResetToken(ctx context.Context, token string) (string, error) {
	key := "reset_password:token:" + token
	cmd := a.client.B().Getdel().Key(key).Build()

	result, err := a.client.Do(ctx, cmd).ToString()
	if err != nil {
		log.Err(err).Msg("failed to fetch reset password token from redis")
		if rueidis.IsRedisNil(err) {
			return "", fmt.Errorf("reset password token is not valid or expierd: %w", err)
		}
		return "", fmt.Errorf("failed to fetch reset password token from redis: %w", err)
	}

	return result, nil
}

func (a *AccountRedisRepository) ByResetEmail(ctx context.Context, email string) (string, error) {
	key := "reset_password:email:" + email
	cmd := a.client.B().Get().Key(key).Build()

	result, err := a.client.Do(ctx, cmd).ToString()
	if err != nil {
		log.Err(err).Str("email", email).Msg("failed to fetch reset password email from redis")
		if rueidis.IsRedisNil(err) {
			return "", fmt.Errorf("reset password email is not valid or expierd: %w", err)
		}
		return "", fmt.Errorf("failed to fetch reset password email from redis: %w", err)
	}

	return result, nil
}
