// Package main grants cmd entree point for whole application
package main

import (
	"github.com/Pklerik/urlshortener/internal/app"
	"github.com/Pklerik/urlshortener/internal/logger"
)

func main() {
	parsedArgs := parseFlags()
	logger.Initialize(parsedArgs.GetLogLevel())
	app.StartApp(parsedArgs)
}
