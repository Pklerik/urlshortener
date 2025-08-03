// Package validators provide base validation logic
package validators

import (
	"net/http"
	"slices"
	"strings"
)

// TextPlain check if Content-Type is `text/plain`.
func TextPlain(res *http.ResponseWriter, req *http.Request) {
	if value, ok := req.Header[`Content-Type`]; !ok {
		http.Error(*res, `Empty content type`, http.StatusBadRequest)
	} else if !slices.ContainsFunc(value, func(s string) bool { return strings.Contains(s, `text/plain`) }) {
		http.Error(*res, `Wrong content type`, http.StatusBadRequest)
	}
}
