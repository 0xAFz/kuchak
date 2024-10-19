package repository

import (
	"context"
	"fmt"
	"kuchak/internal/entity"

	"github.com/jackc/pgx/v5/pgxpool"
)

var _ URL = &URLPostgresRepository{}

type URLPostgresRepository struct {
	session *pgxpool.Pool
}

func NewURLPostgresRepository(session *pgxpool.Pool) *URLPostgresRepository {
	return &URLPostgresRepository{
		session: session,
	}
}

func (u *URLPostgresRepository) ByID(ctx context.Context, ID int) (entity.URL, error) {
	query := `SELECT (id, short_url, original_url, user_id, click_count, expiry_date, created_at)
			  FROM urls
			  WHERE id = $1`

	var url entity.URL
	err := u.session.QueryRow(ctx, query, ID).Scan(&url.ID, &url.ShortURL, &url.OriginalURL, &url.UserID, &url.ClickCount, &url.ExpiryDate, &url.CreatedAt)
	if err != nil {
		return entity.URL{}, fmt.Errorf("failed to get url: %v", err)
	}

	return url, nil
}

func (u *URLPostgresRepository) ByShortURL(ctx context.Context, shortURL string) (entity.URL, error) {
	query := `SELECT (id, short_url, original_url, user_id, click_count, expiry_date, created_at)
			  FROM urls
			  WHERE short_url = $1`

	var url entity.URL
	err := u.session.QueryRow(ctx, query, shortURL).Scan(&url.ID, &url.ShortURL, &url.OriginalURL, &url.UserID, &url.ClickCount, &url.ExpiryDate, &url.CreatedAt)
	if err != nil {
		return entity.URL{}, fmt.Errorf("failed to get url: %v", err)
	}

	return url, nil
}

func (u *URLPostgresRepository) Save(ctx context.Context, url entity.URL) error {
	query := `INSERT INTO urls (short_url, original_url, user_id, expiry_date)
			  VALUES ($1, $2, $3, $4)
			  ON CONFLICT (short_url) DO NOTHING`

	tx, err := u.session.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to start transcation on creating url: %w", err)
	}

	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, query, url.ShortURL, url.OriginalURL, url.UserID, url.ExpiryDate)
	if err != nil {
		return fmt.Errorf("failed to create new url: %w", err)
	}

	err = tx.Commit(ctx)
	if err != nil {
		return fmt.Errorf("failed to create new url: %w", err)
	}

	return nil
}

func (u *URLPostgresRepository) Delete(ctx context.Context, url entity.URL) error {
	query := `DELETE FROM urls
			  WHERE short_url = $1`
	_, err := u.session.Exec(ctx, query, url.ShortURL)
	if err != nil {
		return fmt.Errorf("failed to delete url: %w", err)
	}

	return nil
}

func (u *URLPostgresRepository) UpdateClickCount(ctx context.Context, shortURL string) error {
	query := `UPDATE urls
			  SET click_count = click_count + 1
			  WHERE short_url = $1
			  `
	_, err := u.session.Exec(ctx, query, shortURL)
	if err != nil {
		return fmt.Errorf("failed to increment click count: %w", err)
	}
	return nil
}
