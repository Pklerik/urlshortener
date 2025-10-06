// Package logger provide singleton Log logger for service.
package logger

import (
	"fmt"
	"net/http"

	"go.uber.org/zap"
)

// Log будет доступен всему коду как синглтон.
// По умолчанию установлен no-op-логер, который не выводит никаких сообщений.
var Log *zap.Logger = zap.NewNop()

// Sugar *zap.SugaredLogger.
var Sugar *zap.SugaredLogger

// Initialize инициализирует синглтон логера с необходимым уровнем логирования.
func Initialize(level string) error {
	// преобразуем текстовый уровень логирования в zap.AtomicLevel
	lvl, err := zap.ParseAtomicLevel(level)
	if err != nil {
		return fmt.Errorf("Initialize: %w", err)
	}

	config := zap.Config{
		Level:            lvl,
		Development:      false,
		Encoding:         "json",
		EncoderConfig:    zap.NewDevelopmentEncoderConfig(),
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stderr"},
	}

	// создаём логер на основе конфигурации
	zl, err := config.Build()
	if err != nil {
		return fmt.Errorf("Initialize: %w", err)
	}
	// устанавливаем синглтон
	Log = zl
	Sugar = Log.Sugar()

	return nil
}

// RequestLogger — middleware-логер для входящих HTTP-запросов.
func RequestLogger(h http.HandlerFunc) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		Log.Debug("got incoming HTTP request",
			zap.String("method", r.Method),
			zap.String("path", r.URL.Path),
		)
		h(w, r)
	})
}
