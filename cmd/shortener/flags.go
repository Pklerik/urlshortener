package main

import (
	"flag"
	"log"

	"github.com/Pklerik/urlshortener/internal/config"
	"github.com/Pklerik/urlshortener/internal/config/audit"
	"github.com/Pklerik/urlshortener/internal/config/dbconf"
	"github.com/Pklerik/urlshortener/pkg/random"
	"github.com/caarlos0/env/v11"
)

func parseFlags() config.StartupFlagsParser {
	// Set default vars
	secretKey, err := random.RandBytes(32)
	if err != nil {
		log.Fatal(err)
	}

	parsedArgs := new(config.StartupFlags)
	parsedArgs.ServerAddress = new(config.Address)
	parsedArgs.DBConf = new(dbconf.Conf)
	parsedArgs.Audit = new(audit.Audit)
	flag.Var(parsedArgs.ServerAddress, "a", "address and port to run server")
	flag.StringVar(&parsedArgs.BaseURL, "b", "http://localhost:8080", "protocol://address:port for shortened urls")
	flag.Float64Var(&parsedArgs.Timeout, "timeout", 600, "Custom timeout. Example: --timeout 635.456 sets timeout to 635.456 seconds. Default: 600s")
	flag.StringVar(&parsedArgs.LogLevel, "log_level", "info", "Custom logging level. Default: INFO")
	flag.StringVar(&parsedArgs.LocalStorage, "f", "local_storage.json", "Custom local file location for data storage")
	flag.Var(parsedArgs.DBConf, "d", "Database login DNS string")
	flag.StringVar(&parsedArgs.SecretKey, "secret_key", secretKey, "Secret key for crypto")
	flag.StringVar(&parsedArgs.Audit.LogFilePath, "audit_file", "", "File path for audit log")
	flag.StringVar(&parsedArgs.Audit.LogUrlPath, "audit_url", "", "URL path for audit log")
	flag.Parse()

	envArgs := new(config.StartupFlags)
	envArgs.Audit = new(audit.Audit)

	err = env.Parse(envArgs)
	if err != nil {
		log.Fatal(err)
	}
	err = env.Parse(envArgs.Audit)
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

	if envArgs.DBConf != nil {
		parsedArgs.DBConf = envArgs.DBConf
	}

	if envArgs.SecretKey != "" {
		parsedArgs.SecretKey = envArgs.SecretKey
	} else {
		parsedArgs.SecretKey = secretKey
	}

	if envArgs.Audit.LogFilePath != "" {
		parsedArgs.Audit.LogFilePath = envArgs.Audit.LogFilePath
	}

	if envArgs.Audit.LogUrlPath != "" {
		parsedArgs.Audit.LogUrlPath = envArgs.Audit.LogUrlPath
	}

	return parsedArgs
}
