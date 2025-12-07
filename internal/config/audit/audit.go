// Package audit provide configuration for audit logging.
package audit

import (
	"sync"

	"github.com/go-resty/resty/v2"
)

var (
	onceURL sync.Once
	client  *resty.Client
)

// Audit provide configuration for audit logging.
type Audit struct {
	LogFilePath string `env:"AUDIT_FILE"`
	LogURLPath  string `env:"AUDIT_URL"`
}

// GetLogURLPath provide URL path for audit logging.
func (a *Audit) GetLogURLPath() string {
	return a.LogURLPath
}

// GetLogFilePath provide file path for audit logging.
func (a *Audit) GetLogFilePath() string {
	return a.LogFilePath
}

// GetURLWriter provide resty client for audit logging.
func (a *Audit) GetURLWriter() *resty.Client {
	onceURL.Do( // функция ниже выполнится только один раз
		func() {
			// инициализируем объект
			client = resty.New()
		})

	return client
}
