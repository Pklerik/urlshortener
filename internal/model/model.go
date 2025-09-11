// Package model contains structures for application and doesn't contain business logic.
package model

import "fmt"

// LinkUUIDv7 is a custom type that embeds uuidv7.UUID.
type LinkUUIDv7 string

// LinkData provide structure for URLs storage.
type LinkData struct {
	UUID     LinkUUIDv7 `json:"uuid"`
	ShortURL string     `json:"short_url"`
	LongURL  string     `json:"original_url"`
}

func (ld *LinkData) String() string {
	return fmt.Sprintf(`LinkData{UUID: %s, ShortURL: %s, LongURL: %s}`, ld.UUID, ld.ShortURL, ld.LongURL)
}

// Request provide request for shortener.
type Request struct {
	URL string `json:"url"`
}

// Response provide response for shortener.
type Response struct {
	Result string `json:"result"`
}
