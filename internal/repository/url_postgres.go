package repository

import (
	"context"
	"fmt"
	"kuchak/internal/entity"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/rs/zerolog/log"
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
	query := `SELECT id, short_url, original_url, user_id, click_count, created_at
			  FROM urls
			  WHERE id = $1`

	var url entity.URL
	err := u.session.QueryRow(ctx, query, ID).Scan(&url.ID, &url.ShortURL, &url.OriginalURL, &url.UserID, &url.ClickCount, &url.CreatedAt)
	if err != nil {
		log.Err(err).Int("id", ID).Msg("failed to fetch url by id")
		return entity.URL{}, fmt.Errorf("failed to fetch url by id: %v", err)
	}

	return url, nil
}

func (u *URLPostgresRepository) ByShortURL(ctx context.Context, shortURL string) (entity.URL, error) {
	query := `SELECT id, short_url, original_url, user_id, click_count, created_at
			  FROM urls
			  WHERE short_url = $1`

	var url entity.URL
	err := u.session.QueryRow(ctx, query, shortURL).Scan(&url.ID, &url.ShortURL, &url.OriginalURL, &url.UserID, &url.ClickCount, &url.CreatedAt)
	if err != nil {
		log.Err(err).Str("short_url", shortURL).Msg("failed to fetch url by short_url")
		return entity.URL{}, fmt.Errorf("failed to fetch url by short_url: %v", err)
	}

	return url, nil
}

func (u *URLPostgresRepository) ByUserID(ctx context.Context, userID int) ([]entity.URL, error) {
	query := `SELECT id, short_url, original_url, user_id, click_count, created_at
			  FROM urls
			  WHERE user_id = $1`

	var urls []entity.URL

	rows, err := u.session.Query(ctx, query, userID)
	if err != nil {
		log.Err(err).Int("user_id", userID).Msg("failed to fetch urls by user id")
		return nil, fmt.Errorf("failed to fetch urls by user id: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var url entity.URL
		if err := rows.Scan(&url.ID, &url.ShortURL, &url.OriginalURL, &url.UserID, &url.ClickCount, &url.CreatedAt); err != nil {
			log.Err(err).Msg("failed to scan url row")
			return nil, fmt.Errorf("failed to scan url row: %v", err)
		}
		urls = append(urls, url)
	}

	if err := rows.Err(); err != nil {
		log.Err(err).Msg("failed to iterate url rows")
		return nil, fmt.Errorf("rows iteration error: %v", err)
	}

	return urls, nil
}

func (u *URLPostgresRepository) Save(ctx context.Context, url entity.URL) error {
	query := `INSERT INTO urls (short_url, original_url, user_id)
			  VALUES ($1, $2, $3)
			  ON CONFLICT (short_url) DO NOTHING`

	tx, err := u.session.Begin(ctx)
	if err != nil {
		log.Err(err).Msg("failed to start transcation on creating url")
		return fmt.Errorf("failed to start transcation on creating url: %w", err)
	}

	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, query, url.ShortURL, url.OriginalURL, url.UserID)
	if err != nil {
		log.Err(err).Interface("url", url).Msg("failed to create url")
		return fmt.Errorf("failed to create url: %w", err)
	}

	err = tx.Commit(ctx)
	if err != nil {
		log.Err(err).Msg("failed to commit create url transcation")
		return fmt.Errorf("failed to create url: %w", err)
	}

	return nil
}

func (u *URLPostgresRepository) Delete(ctx context.Context, url entity.URL) error {
	query := `DELETE FROM urls
			  WHERE short_url = $1`
	_, err := u.session.Exec(ctx, query, url.ShortURL)
	if err != nil {
		log.Err(err).Interface("url", url).Msg("failed to delete url")
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
		log.Err(err).Str("short_url", shortURL).Msg("failed to increment click count")
		return fmt.Errorf("failed to increment click count: %w", err)
	}
	return nil
}
