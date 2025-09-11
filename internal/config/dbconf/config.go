// Package dbconf provide database configurations.
package dbconf

import (
	"errors"
	"fmt"
	"strings"
)

var (
	// ErrEmptyDatabaseDSN - DatabaseDSN is empty.
	ErrEmptyDatabaseDSN = errors.New("DatabaseDSN is empty")
	// ErrIncorrectDatabaseURL database URL is incorrect, please use mask: postgresql://[user[:password]@][host][:port]/[database][?parameters].
	ErrIncorrectDatabaseURL = errors.New("database URL is incorrect, please use mask: postgresql://[user[:password]@][host][:port]/[database][?parameters]")
	// ErrNotImplemented DBConfigurer instance is not implemented.
	ErrNotImplemented = errors.New("DBConfigurer instance is not implemented")
)

var (
	// DefaultGooseDrier - default driver if not presented.
	DefaultGooseDrier = "postgres"
)

// Options is alias for db config options.
type Options map[string]string

// DBConfigurer provide methods for db configuration.
type DBConfigurer interface {
	String() string
	UnmarshalText(text []byte) error
	Set(s string) error
	SetDefault() error
	GetConnString() string
	GetOptions() Options
	GetUser() string
	Valid() bool
}

// Conf contain attrs for DB configuration.
type Conf struct {
	User      string
	Password  string
	Host      string
	Port      string
	Database  string
	Options   Options
	RawString string
}

// UnmarshalText provide text unmarshaling for Address string.
func (dbc *Conf) UnmarshalText(text []byte) error {
	err := dbc.Set(string(text))
	if err != nil {
		return fmt.Errorf("UnmarshalText: %w", err)
	}

	return nil
}

// String provide string representation of Conf.
func (dbc *Conf) String() string {
	return fmt.Sprintf("DSN: %s", dbc.RawString)
}

// Set parse Conf from string.
// nolint
func (dbc *Conf) Set(s string) error {
	if len(s) == 0 {
		return nil
	}

	dialectIdx := strings.Index(s, "://")
	if dialectIdx == -1 {
		return ErrIncorrectDatabaseURL
	}

	userIdx := strings.Index(s[dialectIdx+3:], ":")
	if userIdx == -1 {
		return ErrIncorrectDatabaseURL
	}

	userIdx += dialectIdx + 3
	dbc.User = s[dialectIdx+3 : userIdx]

	passwordIdx := strings.Index(s[userIdx:], "@")
	if passwordIdx == -1 {
		return ErrIncorrectDatabaseURL
	}

	passwordIdx += userIdx
	dbc.Password = s[userIdx+1 : passwordIdx]

	hostIdx := strings.Index(s[passwordIdx:], ":")
	if hostIdx == -1 {
		return ErrIncorrectDatabaseURL
	}

	hostIdx += passwordIdx
	dbc.Host = s[passwordIdx+1 : hostIdx]

	portIdx := strings.Index(s[hostIdx:], "/")
	if portIdx == -1 {
		return ErrIncorrectDatabaseURL
	}

	portIdx += hostIdx
	if len(s[hostIdx+1:portIdx]) > 0 {
		dbc.Port = s[hostIdx+1 : portIdx]
	} else {
		dbc.Port = "5432"
	}

	connOptionsIdx := strings.Index(s[portIdx:], "?")
	if connOptionsIdx == -1 {
		dbc.Database = s[portIdx+1:]
		return nil
	}

	connOptionsIdx += portIdx
	dbc.Database = s[portIdx+1 : connOptionsIdx]

	connOptions := strings.Split(s[connOptionsIdx+1:], "?")
	if len(connOptions) == 0 {
		dbc.Options = make(Options, 0)
		return nil
	}

	dbc.Options = make(Options, len(connOptions)-1)
	for i := 1; i < len(connOptions); i++ {
		var name, value string

		name = strings.Split(connOptions[i], "=")[0]
		if len(strings.Split(connOptions[i], "=")) > 1 {
			value = strings.Split(connOptions[i], "=")[1]
		}

		dbc.Options[name] = value
	}

	if err := dbc.SetDefault(); err != nil {
		return fmt.Errorf("unable to set defaults")
	}

	return nil
}

// SetDefault provide default vals for options.
func (dbc *Conf) SetDefault() error {
	if _, ok := dbc.Options["search_path"]; !ok {
		dbc.Options["search_path"] = "public"
	}

	// if _, ok := dbc.Options["sslmode"]; !ok {
	// 	dbc.Options["sslmode"] = ""
	// }

	return nil
}

// GetConnString provide coonection string for db.
func (dbc *Conf) GetConnString() string {
	ps := fmt.Sprintf("user=%s password=%s dbname=%s host=%s port=%s", dbc.User, dbc.Password, dbc.Database, dbc.Host, dbc.Port)
	if dbc.Options == nil {
		return ps
	}

	for name, value := range dbc.Options {
		ps += fmt.Sprintf(" %s=%s", name, value)
	}

	return ps
}

// GetUser returns user.
func (dbc *Conf) GetUser() string {
	return dbc.User
}

// GetOptions return Options.
func (dbc *Conf) GetOptions() Options {
	return dbc.Options
}

// Valid return is (dbc *Conf) is valid config.
func (dbc *Conf) Valid() bool {
	switch {
	case dbc.User == "":
		return false
	case dbc.Database == "":
		return false
	case dbc.Host == "":
		return false
	default:
		return true
	}
}
