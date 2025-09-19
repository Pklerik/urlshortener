// Package dbconf provide database configurations.
package dbconf

import (
	"errors"
	"fmt"
	"regexp"
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
	Dialect   string
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
func (dbc *Conf) Set(s string) error {
	var err error

	if len(s) == 0 {
		return ErrEmptyDatabaseConfig
	}

	dbc.RawString = s

	dbc.Dialect, err = getDialect(s)
	if err != nil {
		return fmt.Errorf("error getting dialect: %w", err)
	}
	credentials, err := getCredentials(s)
	if err != nil {
		return fmt.Errorf("error getting credentials: %w", err)
	}
	dbc.User = credentials.user
	dbc.Password = credentials.pass

	spec, err := getDBSpec(s)
	if err != nil {
		return fmt.Errorf("error getting dbspec: %w", err)
	}
	dbc.Host = spec.host
	dbc.Port = spec.port
	dbc.Database = spec.database

	dbc.Options = getOptions(s)

	if err := dbc.SetDefault(); err != nil {
		return fmt.Errorf("unable to set defaults")
	}

	return nil
}

func getDialect(s string) (string, error) {
	dialectExp := regexp.MustCompile(`.*:\/\/`)
	dialect := dialectExp.FindString(s)
	dialect = strings.Trim(dialect, ":/")
	return dialect, nil
}

func getCredentials(s string) (struct {
	user string
	pass string
}, error) {
	credentialsExp := regexp.MustCompile(`:\/\/.*@`)
	credentials := strings.Trim(credentialsExp.FindString(s), ":/@")
	credentialsSl := strings.Split(credentials, ":")
	if len(credentialsSl) < 2 {
		return struct {
			user string
			pass string
		}{}, ErrIncorrectDatabaseURL
	}

	return struct {
		user string
		pass string
	}{user: credentialsSl[0], pass: credentialsSl[1]}, nil
}

type dbSpec struct {
	host     string
	port     string
	database string
}

func getDBSpec(s string) (dbSpec, error) {
	dbs := dbSpec{}
	bdSpecExp := new(regexp.Regexp)
	if strings.Contains(s, "?") {
		bdSpecExp = regexp.MustCompile(`@.*\?`)
	} else {
		bdSpecExp = regexp.MustCompile(`@.*$`)
	}

	dbStr := bdSpecExp.FindString(s)
	dbStr = strings.Trim(dbStr, "@?")
	if !strings.Contains(dbStr, ":") || !strings.Contains(dbStr, "/") {
		return dbs, ErrIncorrectDatabaseURL
	}
	dbs.host = dbStr[:strings.Index(dbStr, ":")]
	dbs.port = dbStr[strings.Index(dbStr, ":")+1 : strings.Index(dbStr, "/")]
	dbs.database = dbStr[strings.Index(dbStr, "/")+1:]
	return dbs, nil
}

func getOptions(s string) Options {
	optionsIdx := strings.Index(s, "?")
	if optionsIdx == -1 {
		return Options{}
	}
	slOptions := strings.Split(s[optionsIdx+1:], "?")
	options := make(Options, len(slOptions))
	for _, option := range slOptions {
		options[option[:strings.Index(option, "=")]] = option[strings.Index(option, "=")+1:]
	}
	return options
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
