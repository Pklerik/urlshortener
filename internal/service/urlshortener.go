package service

import (
	"crypto/rand"

	"github.com/Pklerik/urlshortener/internal/repository"
)

func ShortURL(long []byte) (short [10]byte, err error) {
	for shortUrl, longUrl := range repository.MapShortener {
		if string(long) == string(longUrl) {
			return shortUrl, nil
		}
	}
	for {
		short = [10]byte([]byte(rand.Text()[:10]))
		if _, ok := repository.MapShortener[short]; !ok {
			repository.MapShortener[short] = long
			break
		}
	}
	return
}
