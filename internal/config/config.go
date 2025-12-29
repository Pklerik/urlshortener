// Package config provide all app configs.
package config

import (
	"fmt"
	"reflect" // nolint:depguard // used for config merging

	"github.com/goccy/go-json"

	"strconv"
	"strings"
	"time"

	"github.com/Pklerik/urlshortener/internal/config/audit"
	"github.com/Pklerik/urlshortener/internal/config/dbconf"
)

// StartupFlagsParser provide interface for app flags.
type StartupFlagsParser interface {
	GetServerAddress() Address
	GetAddressShortURL() string
	GetTimeout() time.Duration
	GetLogLevel() string
	GetLocalStorage() string
	GetDatabaseConf() (dbconf.DBConfigurer, error)
	GetSecretKey() string
	GetAudit() *audit.Audit
	GetTLS() bool
}

// StartupFlags app startup flags.
type StartupFlags struct {
	ServerAddress *Address     `json:"server_address" env:"SERVER_ADDRESS"`
	DBConf        *dbconf.Conf `json:"database_dsn" env:"DATABASE_DSN"`
	Audit         *audit.Audit `json:"audit" env:"AUDIT"`
	BaseURL       string       `json:"base_url" env:"BASE_URL"`
	LogLevel      string       `json:"log_level" env:"LOG_LEVEL"`
	LocalStorage  string       `json:"local_storage_path" env:"FILE_STORAGE_PATH"`
	SecretKey     string       `json:"secret_key" env:"SECRET_KEY"`
	FileConfig    string       `env:"CONFIG"`
	Timeout       float64      `json:"timeout" env:"SERVER_TIMEOUT"`
	TLS           bool         `json:"enable_https" env:"ENABLE_HTTPS"`
}

// GetServerAddress returns ServerAddress.
func (sf *StartupFlags) GetServerAddress() Address {
	return *sf.ServerAddress
}

// GetAddressShortURL returns AddressShortURL.
func (sf *StartupFlags) GetAddressShortURL() string {
	return sf.BaseURL
}

// GetTimeout returns GetTimeout.
func (sf *StartupFlags) GetTimeout() time.Duration {
	return time.Duration(sf.Timeout * float64(time.Second))
}

// GetLogLevel returns LogLevel.
func (sf *StartupFlags) GetLogLevel() string {
	return sf.LogLevel
}

// GetLocalStorage returns LogLevel.
func (sf *StartupFlags) GetLocalStorage() string {
	return sf.LocalStorage
}

// GetSecretKey returns SecretKey.
func (sf *StartupFlags) GetSecretKey() string {
	return sf.SecretKey
}

// GetDatabaseConf returns pointer to dbconf.DBConfigurer or nil.
func (sf *StartupFlags) GetDatabaseConf() (dbconf.DBConfigurer, error) {
	if sf.DBConf == nil {
		return nil, dbconf.ErrEmptyDatabaseConfig
	}

	if err := sf.DBConf.Valid(); err != nil {
		return nil, fmt.Errorf("GetDatabaseConf: %w", err)
	}

	return sf.DBConf, nil
}

// GetAudit returns Audit config.
func (sf *StartupFlags) GetAudit() *audit.Audit {
	return sf.Audit
}

// GetTLS returns TLS.
func (sf *StartupFlags) GetTLS() bool {
	return sf.TLS
}

// Address base struct.
type Address struct {
	Protocol string
	Host     string
	Port     int
}

// UnmarshalText provide text unmarshaling for Address string.
func (a *Address) UnmarshalText(text []byte) error {
	err := a.Set(string(text))
	if err != nil {
		return fmt.Errorf("UnmarshalText: %w", err)
	}

	return nil
}

// String provide string representation of Address.
func (a *Address) String() string {
	if a.Protocol == "" {
		a.Protocol = "http"
	}

	if a.Host == "" {
		a.Host = "localhost"
	}

	if a.Port == 0 {
		a.Port = 8080
	}

	return fmt.Sprintf("%s://%s:%d", a.Protocol, a.Host, a.Port)
}

// Set parse Address from string.
func (a *Address) Set(flagValue string) error {
	flagValueMod := flagValue
	if strings.Contains(flagValue, "http") {
		a.Protocol = strings.Split(flagValue, ":")[0]
	}

	if len(strings.Split(flagValue, "://")) == 2 {
		flagValueMod = strings.Split(flagValue, "://")[1]
	}

	if len(strings.Split(flagValueMod, "/")) > 1 {
		flagValueMod = strings.Split(flagValueMod, "/")[0]
	}

	a.Host = strings.Split(flagValueMod, ":")[0]

	port, err := strconv.Atoi(strings.Split(flagValueMod, ":")[1])
	if err != nil {
		return fmt.Errorf("can't set Address Port for %s: %w", flagValue, err)
	}

	if a.Protocol == "" {
		a.Protocol = "http"
	}

	if a.Host == "" {
		a.Host = "localhost"
	}

	a.Port = port
	if a.Port == 0 {
		a.Port = 8080
	}

	return nil
}

// UnmarshalJSON provide custom unmarshaling for StartupFlags.
func (sf *StartupFlags) UnmarshalJSON(data []byte) error {
	unmarshalMap := make(map[string]interface{}, reflect.TypeOf(*sf).NumField())

	err := json.Unmarshal(data, &unmarshalMap)
	if err != nil {
		return fmt.Errorf("UnmarshalJSON: %w", err)
	}

	for key, value := range unmarshalMap {
		switch key {
		case "timeout":
			if v, ok := value.(float64); ok {
				sf.Timeout = v
			}
		case "server_address":
			if v, ok := value.(string); ok {
				adr := Address{}
				if err := adr.UnmarshalText([]byte(v)); err != nil {
					return fmt.Errorf("UnmarshalJSON server_address: %w", err)
				}

				if sf.ServerAddress == nil {
					sf.ServerAddress = new(Address)
				}

				*(sf.ServerAddress) = adr
			}
		case "db_conf":
			if v, ok := value.(string); ok {
				dbConf := &dbconf.Conf{}

				err := json.Unmarshal([]byte(v), dbConf)
				if err != nil {
					return fmt.Errorf("UnmarshalJSON db_conf: %w", err)
				}

				if sf.DBConf == nil {
					sf.DBConf = new(dbconf.Conf)
				}

				(*sf.DBConf) = (*dbConf)
			}
		case "base_url":
			if v, ok := value.(string); ok {
				sf.BaseURL = v
			}
		case "log_level":
			if v, ok := value.(string); ok {
				sf.LogLevel = v
			}
		case "file_storage_path":
			if v, ok := value.(string); ok {
				sf.LocalStorage = v
			}
		case "secret_key":
			if v, ok := value.(string); ok {
				sf.SecretKey = v
			}
		case "audit":
			if v, ok := value.(string); ok {
				auditConf := &audit.Audit{}

				err := json.Unmarshal([]byte(v), auditConf)
				if err != nil {
					return fmt.Errorf("UnmarshalJSON audit: %w", err)
				}

				sf.Audit = auditConf
			}
		case "enable_https":
			if v, ok := value.(bool); ok {
				sf.TLS = v
			}
		}
	}

	return nil
}
