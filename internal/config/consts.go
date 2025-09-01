// Package config contains all configuration for the app
package config

import (
	"log"
	"os"
	"path/filepath"
)

const (
	// MapSize - base map size.
	MapSize = 100
	// ShortURLLen provide len of short url.
	ShortURLLen = 10
)

var (
	// BasePath provide root dir for project.
	BasePath string
)

func init() {
	configPath, err := os.Getwd()
	if err != nil {
		log.Fatalf("Unable to get PWD: %v", err.Error())
	}

	BasePath = filepath.Dir(filepath.Dir(configPath))
}
