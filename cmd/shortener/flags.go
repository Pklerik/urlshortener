package main

import (
	"flag"

	"github.com/Pklerik/urlshortener/internal/config"
)

func parseFlags() *config.StartupFalgs {
	// Set default vars
	parsedArgs := new(config.StartupFalgs)
	flag.Var(&parsedArgs.ServerAddress, "a", "address and port to run server")
	flag.StringVar(&parsedArgs.AddressShortURL, "b", "http://localhost:8080", "protocol://address:port for shortened urls")
	flag.Var(&parsedArgs.Timeout, "timeout", "Custom timeout. Example: --timeout 1m2s3ms sets timeout to 62.003 seconds. Default: 10m")
	flag.Parse()

	return parsedArgs
}
