package main

import (
	"flag"
	"log"

	"github.com/Pklerik/urlshortener/internal/config"
	"github.com/caarlos0/env/v11"
)

func parseFlags() config.StartupFlagsParser {
	// Set default vars
	parsedArgs := config.NewStartupFlags()
	flag.Var(parsedArgs.ServerAddress, "a", "address and port to run server")
	flag.StringVar(&parsedArgs.BaseURL, "b", "http://localhost:8080", "protocol://address:port for shortened urls")
	flag.Float64Var(&parsedArgs.Timeout, "timeout", 1600, "Custom timeout. Example: --timeout 635.456 sets timeout to 635.456 seconds. Default: 600s")
	flag.StringVar(&parsedArgs.LogLevel, "log_level", "info", "Custom logging level. Default: INFO")
	flag.StringVar(&parsedArgs.LocalStorage, "f", "local_storage.json", "Custom local file location for data storage")
	flag.StringVar(&parsedArgs.DBConf.DatabaseDSN, "d", "", "Database login DNS string")
	flag.Parse()

	envArgs := config.NewStartupFlags()

	err := env.Parse(envArgs)
	if err != nil {
		log.Fatal(err)
	}

	if envArgs.ServerAddress != nil {
		parsedArgs.ServerAddress = envArgs.ServerAddress
	}

	if envArgs.BaseURL != "" {
		parsedArgs.BaseURL = envArgs.BaseURL
	}

	if envArgs.LocalStorage != "" {
		parsedArgs.LocalStorage = envArgs.LocalStorage
	}

	if envArgs.DBConf.DatabaseDSN != "" {
		parsedArgs.DBConf.DatabaseDSN = envArgs.DBConf.DatabaseDSN
	}

	return parsedArgs
}
