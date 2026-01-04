// Package logger provide singleton Log logger for service.
package logger

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"

	"github.com/Pklerik/urlshortener/internal/config/audit"
	"github.com/go-resty/resty/v2"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	// Log будет доступен всему коду как синглтон.
	// По умолчанию установлен no-op-логер, который не выводит никаких сообщений.
	Log *zap.Logger = zap.NewNop()

	// Sugar *zap.SugaredLogger.
	Sugar *zap.SugaredLogger

	config      zap.Config
	auditLogger *zap.Logger
	once        sync.Once
)

// Initialize инициализирует синглтон логера с необходимым уровнем логирования.
func Initialize(level string) error {
	// преобразуем текстовый уровень логирования в zap.AtomicLevel
	lvl, err := zap.ParseAtomicLevel(level)
	if err != nil {
		return fmt.Errorf("Initialize: %w", err)
	}

	config = zap.Config{
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

// GetConfig return zapcore.EncoderConfig.
func GetConfig() zapcore.EncoderConfig {
	return config.EncoderConfig
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

// AuditLogger provide audit logger.
func AuditLogger(auditConf *audit.Audit) *zap.Logger {
	once.Do(func() {
		cores := make([]zapcore.Core, 0, 2)

		emptyCfg := zapcore.EncoderConfig{
			TimeKey:       "",
			LevelKey:      "",
			NameKey:       "",
			CallerKey:     "",
			MessageKey:    "",
			StacktraceKey: "",
		}
		if filepath := auditConf.GetLogFilePath(); filepath != "" {
			cores = append(cores, zapcore.NewCore(
				zapcore.NewJSONEncoder(emptyCfg),      // Custom encoder
				zapcore.AddSync(getLogFile(filepath)), // Sync to stdout
				Log.Level(),                           // Use the same level as the original core
			))
		}

		if auditConf.GetLogURLPath() != "" {
			cores = append(cores, zapcore.NewCore(
				zapcore.NewJSONEncoder(emptyCfg),           // Custom encoder
				zapcore.AddSync(NewAuditClient(auditConf)), // Sync to stdout
				Log.Level(), // Use the same level as the original core
			))
		}

		core := zapcore.NewTee(cores...)
		auditLogger = zap.New(core)
	})

	return auditLogger
}

// AuditClient provide audit client for zapcore.
type AuditClient struct {
	client    *resty.Client
	auditConf *audit.Audit
}

// NewAuditClient create AuditClient.
func NewAuditClient(auditConf *audit.Audit) *AuditClient {
	ac := &AuditClient{
		client:    auditConf.GetURLWriter(),
		auditConf: auditConf,
	}

	return ac
}

// Write implement zapcore.WriteSyncer interface.
func (ac *AuditClient) Write(massage []byte) (int, error) {
	buf := bytes.NewBuffer([]byte{})
	buf.Write(massage)

	resp, err := ac.client.GetClient().Post(ac.auditConf.GetLogURLPath(), "application-json", buf)
	if err != nil {
		Sugar.Errorf("Error sending audit by URL: <%s>: %v", ac.auditConf.GetLogURLPath(), err)
		return 0, fmt.Errorf("unable to send audit log")
	}
	err = resp.Body.Close()
	if err != nil {
		Sugar.Errorf("Error closing response body: %v", err)
	}

	return len(massage), err
}

func getLogFile(filePath string) *os.File {
	var (
		err      error
		fullPath = filePath
	)

	if !strings.HasPrefix(filePath, "/") {
		ex, err := os.Executable()
		if err != nil {
			log.Panicln(err)
		}

		fullPath = filepath.Dir(filepath.Dir(filepath.Dir(ex))) + "/" + filePath
	}

	if err := os.MkdirAll(path.Dir(path.Clean(fullPath)), 0750); err != nil {
		log.Panicln(err)
	}

	auditFile, err := os.OpenFile(path.Clean(fullPath), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		log.Panicln(err)
	}

	Sugar.Infof("Audit log file initialized: %s", fullPath)

	return auditFile
}
