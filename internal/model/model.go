// Package model contains structures for application and doesn't contain business logic.
package model

// LinkData provide structure for URLs storage.
type LinkData struct {
	ShortURL string
	LongURL  string
}

// Request provide request for shortener.
type Request struct {
	URL string `json:"url"`
}

// Response provide response for shortener.
type Response struct {
	Result string `json:"result"`
}
