// Package validators provide base validation logic
package validators

import (
	"errors"
	"net/http"
	"slices"
	"strings"

	"github.com/Pklerik/urlshortener/internal/logger"
)

var (
	// ErrEmptyContentType - Empty content type.
	ErrEmptyContentType = errors.New("empty content type")
	// ErrWrongContentType - Wrong content type.
	ErrWrongContentType = errors.New("wrong content type")
)

// TextPlain check if Content-Type is `text/plain`.
func TextPlain(res http.ResponseWriter, req *http.Request) error {
	if value, ok := req.Header[`Content-Type`]; !ok {
		http.Error(res, `Empty content type`, http.StatusBadRequest)
		logger.Sugar.Debugf("Content-Encoding :", res.Header().Get("Content-Encoding"))

		return ErrEmptyContentType
	} else if !slices.ContainsFunc(value, func(s string) bool { return strings.Contains(s, `text/plain`) }) {
		http.Error(res, `Wrong content type`, http.StatusBadRequest)
		logger.Sugar.Debugf("Content-Encoding :", res.Header().Get("Content-Encoding"))

		return ErrWrongContentType
	}

	return nil
}

// ApplicationJSON check if Content-Type is `application/json`.
func ApplicationJSON(res http.ResponseWriter, req *http.Request) error {
	if value, ok := req.Header[`Content-Type`]; !ok {
		logger.Sugar.Debugf("Content-Encoding before error:", res.Header().Get("Content-Encoding"))
		http.Error(res, `Empty content type`, http.StatusBadRequest)
		logger.Sugar.Debugf("Content-Encoding after error :", res.Header().Get("Content-Encoding"))

		return ErrEmptyContentType
	} else if !slices.ContainsFunc(value, func(s string) bool { return strings.Contains(s, `application/json`) }) {
		http.Error(res, `Wrong content type`, http.StatusBadRequest)
		logger.Sugar.Debugf("Content-Encoding :", res.Header().Get("Content-Encoding"))

		return ErrWrongContentType
	}

	return nil
}
