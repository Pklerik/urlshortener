// Package main grants cmd entree point for whole application
package main

import (
	"log"

	"github.com/Pklerik/urlshortener/internal/app"
	"github.com/Pklerik/urlshortener/internal/logger"
)

var err error

func main() {
	parsedArgs := parseFlags()

	err = logger.Initialize(parsedArgs.GetLogLevel())
	if err != nil {
		log.Fatalf("Unable to setup logger: main: %s", err.Error())
	}

	app.StartApp(parsedArgs)
}
