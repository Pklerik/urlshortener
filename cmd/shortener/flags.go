package main

import (
	"flag"
	"log"
	"os"
	"path/filepath"
	"reflect" // nolint:depguard // used for dynamic field access
	"strings"

	"github.com/goccy/go-json"

	"github.com/Pklerik/urlshortener/internal/config"
	"github.com/Pklerik/urlshortener/internal/config/audit"
	"github.com/Pklerik/urlshortener/internal/config/dbconf"
	"github.com/Pklerik/urlshortener/internal/dictionary"
	"github.com/caarlos0/env/v11"
)

func parseArgs() config.StartupFlagsParser {
	parsedArgs := parseFlags()
	envArgs := parseEnvs()

	if envArgs.FileConfig != "" {
		parsedArgs.FileConfig = envArgs.FileConfig
	}

	configArgs := parseConfig(parsedArgs.FileConfig)

	for field := range reflect.TypeOf(envArgs).Elem().NumField() {
		// check if env variable is set
		v := reflect.ValueOf(envArgs).Elem().Field(field)
		if !v.IsZero() {
			// replaces parsedArgs field with env variable
			reflect.ValueOf(parsedArgs).Elem().Field(field).Set(v)
		} else {
			// check if  parsedArgs variable is set
			v := reflect.ValueOf(parsedArgs).Elem().Field(field)
			if v.IsZero() {
				// replaces parsedArgs field with configArgs variable
				v.Set(reflect.ValueOf(configArgs).Elem().Field(field))
			}
		}
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
	flag.BoolVar(&parsedArgs.TLS, "s", false, "use tls Listener ")
	flag.StringVar(&parsedArgs.FileConfig, "c", "", "path to config json file")
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

func parseConfig(fileConfig string) *config.StartupFlags {
	configArgs := new(config.StartupFlags)

	fullPath := fileConfig
	if !strings.HasPrefix(fileConfig, "/") {
		fullPath = filepath.Clean(filepath.Join(dictionary.BasePath, fileConfig))
	}

	if fileConfig != "" {
		data, err := os.ReadFile(filepath.Clean(fullPath))
		if err != nil {
			return configArgs
		}

		if err := json.Unmarshal(data, configArgs); err != nil {
			return configArgs
		}
	}

	return configArgs
}
