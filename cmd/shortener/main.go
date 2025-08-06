// Package main grants cmd entree point for whole application
package main

import (
	"github.com/Pklerik/urlshortener/internal/app"
)

func main() {
	parsedArgs := parseFlags()
	app.StartApp(parsedArgs)
}
