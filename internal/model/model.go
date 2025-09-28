// Package model contains structures for application and doesn't contain business logic.
package model

import (
	"fmt"
)

// LinkUUIDv7 is a custom type that embeds uuidv7.UUID.
type LinkUUIDv7 string

// UserID is a custom type for user id.
type UserID int

// LinkData provide structure for URLs storage.
type LinkData struct {
	UUID     LinkUUIDv7 `json:"uuid"`
	ShortURL string     `json:"short_url"`
	LongURL  string     `json:"original_url"`
	UserID   UserID     `json:"user_id"`
}

func (ld *LinkData) String() string {
	return fmt.Sprintf(`LinkData{UUID: %s, ShortURL: %s, LongURL: %s, UserId: %d}`, ld.UUID, ld.ShortURL, ld.LongURL, ld.UserID)
}
