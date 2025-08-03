// Package service provide all business logic for applications.
package service

import (
	"crypto/rand"
	"log"

	"github.com/Pklerik/urlshortener/internal/repository"
)

// ShortURL grands short url from db.
func ShortURL(long []byte) (string, error) {
	var short string
	for shortURL, longURL := range *repository.MapShorts() {
		if string(long) == longURL {
			log.Printf("Short url: %s get for long: %s", short, string(long))

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

	log.Printf("Short url: %s sets for long: %s", short, string(long))

	return short, nil
}
