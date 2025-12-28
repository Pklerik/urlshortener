package main

import (
	"flag"
	"log"

	"github.com/Pklerik/urlshortener/internal/config"
	"github.com/Pklerik/urlshortener/internal/config/audit"
	"github.com/Pklerik/urlshortener/internal/config/dbconf"
	"github.com/caarlos0/env/v11"
)

func parseArgs() config.StartupFlagsParser {
	parsedArgs := parseFlags()
	envArgs := parseEnvs()

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
	}

	if envArgs.Audit.LogFilePath != "" {
		parsedArgs.Audit.LogFilePath = envArgs.Audit.LogFilePath
	}

	if envArgs.Audit.LogURLPath != "" {
		parsedArgs.Audit.LogURLPath = envArgs.Audit.LogURLPath
	}

	if envArgs.Tls {
		parsedArgs.Tls = envArgs.Tls
	}

	return parsedArgs
}

func parseFlags() *config.StartupFlags {
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
	flag.StringVar(&parsedArgs.SecretKey, "secret_key", "secret_key", "Secret key for crypto")
	flag.StringVar(&parsedArgs.Audit.LogFilePath, "audit_file", "", "File path for audit log")
	flag.StringVar(&parsedArgs.Audit.LogURLPath, "audit_url", "", "URL path for audit log")
	flag.BoolVar(&parsedArgs.Tls, "s", false, "use tls Listener ")
	flag.Parse()

	return parsedArgs
}

func parseEnvs() *config.StartupFlags {
	envArgs := new(config.StartupFlags)
	envArgs.Audit = new(audit.Audit)

	err := env.Parse(envArgs)
	if err != nil {
		log.Fatal(err)
	}

	return envArgs
}
