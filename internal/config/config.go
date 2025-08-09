package config

import (
	"fmt"
	"strconv"
	"strings"
)

// StartupFlagsParser provide interface for app flags.
type StartupFlagsParser interface {
	GetFlags() string
	GetServerAddress() Address
	GetAddressShortURL() string
	GetTimeout() float64
}

// StartupFlags app startup flags.
type StartupFlags struct {
	ServerAddress   Address
	AddressShortURL string
	Timeout         float64
}

// GetFlags provide string representation of flag.
func (sf *StartupFlags) GetFlags() string {
	return fmt.Sprintf("ServerAddress: %s, AddressShortURL: %s, Timeout: %f",
		sf.ServerAddress.String(),
		sf.AddressShortURL,
		sf.Timeout,
	)
}

// GetServerAddress returns ServerAddress.
func (sf *StartupFlags) GetServerAddress() Address {
	return sf.ServerAddress
}

// GetAddressShortURL returns AddressShortURL.
func (sf *StartupFlags) GetAddressShortURL() string {
	return sf.AddressShortURL
}

// GetTimeout returns GetTimeout.
func (sf *StartupFlags) GetTimeout() float64 {
	return sf.Timeout
}

// Address base struct.
type Address struct {
	Protocol string
	Host     string
	Port     int
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
