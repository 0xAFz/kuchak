package entity

import "time"

type URL struct {
	ID          int        `json:"id"`
	ShortURL    string     `json:"short_url"`
	OriginalURL string     `json:"original_url"`
	UserID      int        `json:"user_id"`
	ClickCount  int        `json:"click_count"`
	ExpiryDate  *time.Time `json:"expiry_date,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
}
