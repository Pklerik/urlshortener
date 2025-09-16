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
	// ErrEmptyDatabaseConfig Conf is empty.
	ErrEmptyDatabaseConfig = errors.New("Conf is empty")
)

var (
	// DefaultGooseDrier - default driver if not presented.
	DefaultGooseDrier = "postgres"
)

// ErrNotValidDBConf config fields is not valid.
type ErrNotValidDBConf struct {
	fields []string
}

func (err ErrNotValidDBConf) Error() string {
	return fmt.Sprintf("config field is not valid: %v", err.fields)
}

// Options is alias for db config options.
type Options map[string]string

/*
DBConfigurer - interface for DB configuration.
*/
type DBConfigurer interface {
	SetDefault() error
	GetConnString() string
	GetOptions() Options
	GetUser() string
	Valid() error
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
	if err != nil && !errors.Is(err, ErrEmptyDatabaseDSN) {
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
		return ErrEmptyDatabaseConfig
	}

	dbc.RawString = s

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
		connOptionsIdx = portIdx
	} else {
		connOptionsIdx += portIdx
		dbc.Database = s[portIdx+1 : connOptionsIdx]
	}

	connOptions := strings.Split(s[connOptionsIdx+1:], "?")

	dbc.Options = make(Options, len(connOptions))
	for i := 0; i < len(connOptions); i++ {
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
func (dbc *Conf) Valid() error {
	var err = ErrNotValidDBConf{
		// Количество полей равно количеству таковых а Conf.
		fields: make([]string, 0, 4),
	}

	switch {
	case dbc.User == "":
		err.fields = append(err.fields, "User")
	case dbc.Database == "":
		err.fields = append(err.fields, "Database")
	case dbc.Host == "":
		err.fields = append(err.fields, "Host")
	case dbc.Port == "":
		err.fields = append(err.fields, "Port")
		fallthrough
	default:
		return nil
	}

	if len(err.fields) != 0 {
		return err
	}

	return nil
}
