// Package validators provide base validation logic
package validators

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTextPlain(t *testing.T) {
	type args struct {
		contentTypes []string
	}
	type want struct {
		code     int
		response string
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{name: "Wrong content", args: args{contentTypes: []string{"text/json"}}, want: want{code: http.StatusBadRequest, response: "Wrong content type\n"}},
		{name: "Empty content", args: args{contentTypes: []string{}}, want: want{code: http.StatusBadRequest, response: "Empty content type\n"}},
		{name: "TextPlain", args: args{contentTypes: []string{"text/plain"}}, want: want{code: http.StatusOK, response: ""}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodGet, "/", nil)
			for _, contentType := range tt.args.contentTypes {
				request.Header.Add(`Content-Type`, contentType)
			}

			w := httptest.NewRecorder()
			TextPlain(w, request)
			res := w.Result()

			assert.Equal(t, tt.want.code, res.StatusCode)

			defer res.Body.Close()
			resBody, err := io.ReadAll(res.Body)
			require.NoError(t, err)
			assert.Equal(t, tt.want.response, string(resBody))

		})
	}
}
