package main

import (
	"flag"
	"log"

	"github.com/Pklerik/urlshortener/internal/config"
	"github.com/caarlos0/env/v11"
)

func parseFlags() config.StartupFlagsParser {
	// Set default vars
	parsedArgs := new(config.StartupFlags)
	flag.Var(parsedArgs.ServerAddress, "a", "address and port to run server")
	flag.StringVar(&parsedArgs.BaseURL, "b", "http://localhost:8080", "protocol://address:port for shortened urls")
	flag.Float64Var(&parsedArgs.Timeout, "timeout", 600, "Custom timeout. Example: --timeout 635.456 sets timeout to 635.456 seconds. Default: 600s")
	flag.Parse()

	envArgs := new(config.StartupFlags)
	err := env.Parse(&envArgs)

	if err != nil {
		log.Fatal(err)
	}

	if envArgs.ServerAddress != nil {
		parsedArgs.ServerAddress = envArgs.ServerAddress
	}

	if envArgs.BaseURL != "" {
		parsedArgs.BaseURL = envArgs.BaseURL
	}

	return parsedArgs
}
