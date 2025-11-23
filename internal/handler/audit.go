package handler

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/Pklerik/urlshortener/internal/config"
	"github.com/Pklerik/urlshortener/internal/logger"
	"github.com/Pklerik/urlshortener/internal/model"
	"go.uber.org/zap"
)

// IAuditor provide audit logging for requests.
type IAuditor interface {
	AuditMiddleware(next http.Handler) http.Handler
}

// Auditor provide audit logging for requests.
type Auditor struct {
	Args config.StartupFlagsParser
	ah   IAuthentication
}

// NewAuditor provide audit logging for requests.
func NewAuditor(args config.StartupFlagsParser, ah IAuthentication) *Auditor {
	return &Auditor{
		Args: args,
		ah:   ah,
	}
}

// AuditMiddleware provide audit logging for requests.
func (a *Auditor) AuditMiddleware(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		var (
			action = "follow"
			req    model.Request
		)

		if a.Args.GetAudit() == nil || (a.Args.GetAudit().GetLogFilePath() == "" && a.Args.GetAudit().LogURLPath == "") {
			next.ServeHTTP(w, r)
			return
		}

		switch {
		case strings.HasSuffix(r.URL.Path, "/"):
			action = "shorten"
		case strings.HasSuffix(r.URL.Path, "/api/shorten"):
			action = "shorten"
		}

		userID, err := a.ah.GetUserIDFromCookie(r)
		if err != nil {
			userID = model.UserID("unauthorized")
		}

		var buf bytes.Buffer

		tee := io.TeeReader(r.Body, &buf)

		body, err := io.ReadAll(tee)
		if err != nil && !errors.Is(err, io.EOF) {
			logger.Log.Debug("cannot read body", zap.Error(err))
		}

		if err := readReq(r, body, &req); err != nil {
			logger.Log.Debug("cannot read request", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
		}

		if req.URL == "" {
			logger.Log.Error("request is empty")
		}

		extendedLogger := logger.AuditLogger(a.Args.GetAudit())
		if extendedLogger == nil {
			next.ServeHTTP(w, r)
			return
		}

		extendedLogger.Log(logger.Log.Level(), "", zap.Int64("ts", time.Now().Unix()),
			zap.String("action", action),
			zap.String("user_id", string(userID)),
			zap.String("url", req.URL),
		)

		next.ServeHTTP(w, r)
	}

	return http.HandlerFunc(fn)
}
