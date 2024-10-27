package repository

import (
	"context"
	"errors"
	"fmt"
	"kuchak/internal/entity"
	"log"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
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
		if errors.Is(err, pgx.ErrNoRows) {
			return entity.User{}, fmt.Errorf("user not found: %w", pgx.ErrNoRows)
		}
		return entity.User{}, fmt.Errorf("db query failed: %w", err)
	}

	return user, nil
}

func (a *AccountPostgresRepository) ByEmail(ctx context.Context, email string) (entity.User, error) {
	query := `SELECT id, email, password, is_email_verified, created_at FROM users WHERE email = $1`

	var user entity.User
	err := a.session.QueryRow(ctx, query, email).Scan(&user.ID, &user.Email, &user.Password, &user.IsEmailVerified, &user.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return entity.User{}, fmt.Errorf("user not found: %w", pgx.ErrNoRows)
		}
		return entity.User{}, fmt.Errorf("db query failed: %w", err)
	}

	return user, nil
}

func (a *AccountPostgresRepository) Save(ctx context.Context, user entity.User) error {
	query := `INSERT INTO users (email, password)
			  VALUES ($1, $2)`

	_, err := a.session.Exec(ctx, query, user.Email, user.Password)
	if err != nil {
		return fmt.Errorf("failed to create new user: %w", err)
	}
	log.Printf("New user created")

	return nil
}

func (a *AccountPostgresRepository) Delete(ctx context.Context, user entity.User) error {
	query := `DELETE FROM users
			  WHERE email = $1`

	_, err := a.session.Exec(ctx, query, user.Email)
	if err != nil {
		return fmt.Errorf("failed to delete user %v, error: %w", user, err)
	}

	return nil
}

func (a *AccountPostgresRepository) UpdateEmail(ctx context.Context, user entity.User) error {
	query := `UPDATE users
			  SET email = $1, is_email_verifed = false
			  WHERE id = $2`

	_, err := a.session.Exec(ctx, query, user.Email, user.ID)
	if err != nil {
		return fmt.Errorf("failed to update email for user: %v, error:  %w", user, err)
	}

	return nil
}

func (a *AccountPostgresRepository) UpdatePassword(ctx context.Context, user entity.User) error {
	query := `UPDATE users
			  SET password = $1
			  WHERE id = $2`

	_, err := a.session.Exec(ctx, query, user.Password, user.ID)
	if err != nil {
		return fmt.Errorf("failed to update password for user: %v, error: %w", user, err)
	}

	return nil
}

func (a *AccountPostgresRepository) UpdateVerifyEmail(ctx context.Context, user entity.User) error {
	query := `UPDATE users
			  SET is_email_verified = $1
			  WHERE id = $2`

	_, err := a.session.Exec(ctx, query, user.IsEmailVerified, user.ID)
	if err != nil {
		return fmt.Errorf("failed to update email verify for user: %v error: %w", user, err)
	}

	return nil
}
