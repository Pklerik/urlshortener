// Package validators provide base validation logic
package validators

import (
	"errors"
	"net/http"
	"slices"
	"strings"
)

var (
	// ErrEmptyContentType - Empty content type
	ErrEmptyContentType = errors.New("Empty content type")
	// ErrWrongContentType - Wrong content type
	ErrWrongContentType = errors.New("Wrong content type")
)

// TextPlain check if Content-Type is `text/plain`.
func TextPlain(res http.ResponseWriter, req *http.Request) error {
	if value, ok := req.Header[`Content-Type`]; !ok {
		http.Error(res, `Empty content type`, http.StatusBadRequest)
		res.WriteHeader(http.StatusBadRequest)
		return ErrEmptyContentType
	} else if !slices.ContainsFunc(value, func(s string) bool { return strings.Contains(s, `text/plain`) }) {
		http.Error(res, `Wrong content type`, http.StatusBadRequest)
		res.WriteHeader(http.StatusBadRequest)
		return ErrWrongContentType
	}
	return nil
}
