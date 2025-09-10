// Package db provide database configurations.
package db

// Conf contain attrs for DB configuration.
type Conf struct {
	DatabaseDSN string `env:"DATABASE_DSN"`
}
