package random

import (
	"crypto/rand"
	"encoding/base64"
)

// Provide simple random slice byte of provided size in string base64 encoding.
func RandBytes(size int) (string, error) {
	b := make([]byte, size)
	_, err := rand.Read(b) // записываем байты в слайс b
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(b), nil
}
