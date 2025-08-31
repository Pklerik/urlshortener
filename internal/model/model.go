// Package model contains structures for application and doesn't contain business logic.
package model

// LinkData provide structure for URLs storage.
type LinkData struct {
	ShortURL string
	LongURL  string
}
type Request struct {
	URL string `json:"url"`
}

type Response struct {
	Result string `json:"result"`
}
