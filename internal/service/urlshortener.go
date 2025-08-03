// Package service provide all business logic for applications.
package service

import (
	"crypto/rand"

	"github.com/Pklerik/urlshortener/internal/repository"
)

// ShortURL grands short url from db.
func ShortURL(long []byte) (string, error) {
	var short string
	for shortURL, longURL := range *repository.MapShorts() {
		if string(long) == longURL {
			return shortURL, nil
		}
	}

	for {
		short = rand.Text()[:10]
		if _, ok := (*repository.MapShorts())[short]; !ok {
			(*repository.MapShorts())[short] = string(long)

			break
		}
	}

	return short, nil
}
