// Package model contains structures for application and doesn't contain business logic.
package model

import (
	"fmt"
)

// UUIDv7 is a custom type that embeds uuidv7.UUID.
type UUIDv7 string

// UserID is a custom type for user id.
type UserID UUIDv7

// LinkData provide structure for URLs storage.
type LinkData struct {
	UUID      UUIDv7 `json:"uuid" db:"uuid"`
	ShortURL  string `json:"short_url" db:"short_url"`
	LongURL   string `json:"original_url" db:"original_url"`
	UserID    UserID `json:"user_id" db:"user_id"`
	IsDeleted bool   `json:"is_deleted" db:"is_deleted"`
}

// User represents the core business model for our app.
type User struct {
	ID UserID `json:"id" db:"id"`
}

func (ld *LinkData) String() string {
	return fmt.Sprintf(`LinkData{UUID: %s, ShortURL: %s, LongURL: %s, UserId: %s}`, ld.UUID, ld.ShortURL, ld.LongURL, ld.UserID)
}
