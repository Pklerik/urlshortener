// Package config provide all app configs.
package config

import (
	"fmt"
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
}

// StartupFlags app startup flags.
type StartupFlags struct {
	ServerAddress *Address     `env:"SERVER_ADDRESS"`
	BaseURL       string       `env:"BASE_URL"`
	LogLevel      string       `env:"LOG_LEVEL"`
	LocalStorage  string       `env:"FILE_STORAGE_PATH"`
	DBConf        *dbconf.Conf `env:"DATABASE_DSN"`
	Timeout       float64
	SecretKey     string `env:"SECRET_KEY"`
	Audit         *audit.Audit
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

func (sf *StartupFlags) GetAudit() *audit.Audit {
	return sf.Audit
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
