package main

import (
	"flag"

	"github.com/Pklerik/urlshortener/internal/config"
)

func parseFlags() *config.StartupFalgs {
	// Set default vars
	parsedArgs := &config.StartupFalgs{
		ServerAddress: config.Address{
			Host: "",
			Port: 8080,
		},
		AddressShortURL: config.Address{
			Protocol: "http",
			Host:     "localhost",
			Port:     8080,
		},
		Timeout: config.Timeout{
			Seconds: 10 * 60,
		},
	}
	flag.Var(&parsedArgs.ServerAddress, "a", "address and port to run server")
	flag.Var(&parsedArgs.AddressShortURL, "b", "protocol, address and port for shortened urls")
	flag.Var(&parsedArgs.Timeout, "timeout", "Custom timeout. Example: --timeout 1m2s3ms sets timeout to 62.003 seconds. Default: 10m")
	flag.Parse()

	return parsedArgs
}
