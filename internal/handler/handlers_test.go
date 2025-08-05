// Package handler contains handling logic for all pages
package handler

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Pklerik/urlshortener/internal/repository"
	"github.com/Pklerik/urlshortener/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func handler() (http.Handler, *LinkHandler) {
	linksRepo := repository.NewInMemoryLinksRepository()
	linksService := service.NewLinksService(linksRepo)
	linksHandler := NewLinkHandler(linksService)
	mux := http.NewServeMux()
	mux.HandleFunc(`/`, linksHandler.RegisterLinkHandler)
	return mux, linksHandler
}

func TestRegisterLinkHandler(t *testing.T) {
	type args struct {
		method        string
		contentTypes  []string
		body          *strings.Reader
		additionalURL string
	}
	type want struct {
		code     int
		response string
		Location string
	}

	mux, linksHandler := handler()
	srv := httptest.NewServer(mux)
	defer srv.Close()
	testURL := "http://ya.ru"
	request := httptest.NewRequest(http.MethodPost, srv.URL, strings.NewReader(testURL))
	request.Header.Add(`Content-Type`, "text/plain")
	w := httptest.NewRecorder()
	linksHandler.RegisterLinkHandler(w, request)
	resPost := w.Result()

	defer resPost.Body.Close()
	resBody, _ := io.ReadAll(resPost.Body)

	tests := []struct {
		name string
		args args
		want want
	}{
		{name: "Post Created", args: args{method: http.MethodPost, contentTypes: []string{"text/plain"}, body: strings.NewReader(testURL)}, want: want{code: http.StatusCreated, response: srv.URL}},
		{name: "Wrong content", args: args{method: http.MethodPost, contentTypes: []string{"application/json"}, body: strings.NewReader(testURL)}, want: want{code: http.StatusBadRequest, response: "Wrong content type\n"}},
		{name: "Empty content", args: args{method: http.MethodPost, body: strings.NewReader(testURL)}, want: want{code: http.StatusBadRequest, response: "Empty content type\n"}},
		{name: "Wrong Redirect", args: args{method: http.MethodGet, additionalURL: "/WERTADSD"}, want: want{code: http.StatusBadRequest, response: "Unable to find long URL for short\n"}},
		{name: "Get Redirect", args: args{method: http.MethodGet, additionalURL: "/" + strings.Split(string(resBody), "/")[3]}, want: want{code: http.StatusTemporaryRedirect, Location: testURL}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.args.body == nil {
				tt.args.body = strings.NewReader("")
			}
			request := httptest.NewRequest(tt.args.method, srv.URL+tt.args.additionalURL, tt.args.body)
			for _, contentType := range tt.args.contentTypes {
				request.Header.Add(`Content-Type`, contentType)
			}

			w := httptest.NewRecorder()
			linksHandler.RegisterLinkHandler(w, request)
			res := w.Result()

			assert.Equal(t, tt.want.code, res.StatusCode)

			defer res.Body.Close()
			resBody, err := io.ReadAll(res.Body)
			require.NoError(t, err)
			assert.Contains(t, string(resBody), tt.want.response)
			if res.Header.Get("Location") != "" {
				assert.Contains(t, res.Header.Values("Location"), tt.want.Location)
			}

		})
	}
}
