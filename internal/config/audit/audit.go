package audit

import (
	"sync"

	"github.com/go-resty/resty/v2"
)

var (
	onceURL sync.Once
	client  *resty.Client
)

type Audit struct {
	LogFilePath string `env:"AUDIT_FILE"`
	LogUrlPath  string `env:"AUDIT_URL"`
}

func (a *Audit) GetLogUrlPath() string {
	return a.LogUrlPath
}

func (a *Audit) GetLogFilePath() string {
	return a.LogFilePath
}

func (a *Audit) GetURLWriter() *resty.Client {
	onceURL.Do( // функция ниже выполнится только один раз
		func() {
			// инициализируем объект
			client = resty.New()
		})
	return client
}
