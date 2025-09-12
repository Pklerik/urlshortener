// Package dictionary contains all configuration for the app
package dictionary

import (
	"log"
	"os"
)

const (
	// MapSize - base map size.
	MapSize = 100
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

	BasePath = configPath
}
