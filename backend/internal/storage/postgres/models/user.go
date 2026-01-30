package models

import (
	"time"

	"github.com/google/uuid"
)

// User represents a player account
type User struct {
	ID                 uuid.UUID  `db:"id" json:"id"`
	TelegramID         int64      `db:"telegram_id" json:"telegram_id"`
	TelegramUsername   *string    `db:"telegram_username" json:"telegram_username,omitempty"`
	TelegramFirstName  string     `db:"telegram_first_name" json:"telegram_first_name"`
	TelegramLastName   *string    `db:"telegram_last_name" json:"telegram_last_name,omitempty"`
	CreatedAt          time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt          time.Time  `db:"updated_at" json:"updated_at"`
}