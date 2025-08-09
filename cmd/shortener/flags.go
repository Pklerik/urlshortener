package main

import (
	"flag"

	"github.com/Pklerik/urlshortener/internal/config"
)

func parseFlags() config.StartupFlagsParser {
	// Set default vars
	parsedArgs := new(config.StartupFlags)
	flag.Var(&parsedArgs.ServerAddress, "a", "address and port to run server")
	flag.StringVar(&parsedArgs.AddressShortURL, "b", "http://localhost:8080", "protocol://address:port for shortened urls")
	flag.Float64Var(&parsedArgs.Timeout, "timeout", 600, "Custom timeout. Example: --timeout 635.456 sets timeout to 635.456 seconds. Default: 600s")
	flag.Parse()

	return parsedArgs
}
