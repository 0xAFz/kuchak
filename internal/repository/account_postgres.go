package repository

import (
	"context"
	"errors"
	"fmt"
	"kuchak/internal/entity"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"
)

var _ Account = &AccountPostgresRepository{}

type AccountPostgresRepository struct {
	session *pgxpool.Pool
}

func NewAccountPostgresRepository(session *pgxpool.Pool) *AccountPostgresRepository {
	return &AccountPostgresRepository{
		session: session,
	}
}

func (a *AccountPostgresRepository) ByID(ctx context.Context, ID int) (entity.User, error) {
	query := `SELECT id, email, password, is_email_verified, created_at FROM users WHERE id = $1`
	var user entity.User
	err := a.session.QueryRow(ctx, query, ID).Scan(&user.ID, &user.Email, &user.Password, &user.IsEmailVerified, &user.CreatedAt)
	if err != nil {
		log.Err(err).Int("id", ID).Msg("failed to fetch user by id")
		if errors.Is(err, pgx.ErrNoRows) {
			return entity.User{}, fmt.Errorf("user not found: %w", pgx.ErrNoRows)
		}
		return entity.User{}, fmt.Errorf("failed to fetch user from db: %w", err)
	}

	return user, nil
}

func (a *AccountPostgresRepository) ByEmail(ctx context.Context, email string) (entity.User, error) {
	query := `SELECT id, email, password, is_email_verified, created_at FROM users WHERE email = $1`

	var user entity.User
	err := a.session.QueryRow(ctx, query, email).Scan(&user.ID, &user.Email, &user.Password, &user.IsEmailVerified, &user.CreatedAt)
	if err != nil {
		log.Err(err).Str("email", email).Msg("failed to fetch user by email")
		if errors.Is(err, pgx.ErrNoRows) {
			return entity.User{}, fmt.Errorf("user not found: %w", pgx.ErrNoRows)
		}
		return entity.User{}, fmt.Errorf("failed to fetch user from db: %w", err)
	}

	return user, nil
}

func (a *AccountPostgresRepository) Save(ctx context.Context, user entity.User) error {
	query := `INSERT INTO users (email, password)
			  VALUES ($1, $2)`

	_, err := a.session.Exec(ctx, query, user.Email, user.Password)
	if err != nil {
		log.Err(err).Interface("user", user).Msg("failed to create user")
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

func (a *AccountPostgresRepository) Delete(ctx context.Context, user entity.User) error {
	query := `DELETE FROM users
			  WHERE email = $1`

	_, err := a.session.Exec(ctx, query, user.Email)
	if err != nil {
		log.Err(err).Interface("user", user).Msg("failed to delete user")
		return fmt.Errorf("failed to delete user: %w", err)
	}

	return nil
}

func (a *AccountPostgresRepository) UpdateEmail(ctx context.Context, user entity.User) error {
	query := `UPDATE users
			  SET email = $1, is_email_verifed = false
			  WHERE id = $2`

	_, err := a.session.Exec(ctx, query, user.Email, user.ID)
	if err != nil {
		log.Err(err).Interface("user", user).Msg("failed to update email")
		return fmt.Errorf("failed to update email: %w", err)
	}

	return nil
}

func (a *AccountPostgresRepository) UpdatePassword(ctx context.Context, user entity.User) error {
	query := `UPDATE users
			  SET password = $1
			  WHERE id = $2`

	_, err := a.session.Exec(ctx, query, user.Password, user.ID)
	if err != nil {
		log.Err(err).Interface("user", user).Msg("failed to update password")
		return fmt.Errorf("failed to update password: %w", err)
	}

	return nil
}

func (a *AccountPostgresRepository) UpdateVerifyEmail(ctx context.Context, user entity.User) error {
	query := `UPDATE users
			  SET is_email_verified = $1
			  WHERE id = $2`

	_, err := a.session.Exec(ctx, query, user.IsEmailVerified, user.ID)
	if err != nil {
		log.Err(err).Interface("user", user).Msg("failed to update email verification")
		return fmt.Errorf("failed to update email verification: %w", err)
	}

	return nil
}
