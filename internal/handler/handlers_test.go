// Package handler contains handling logic for all pages
package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Pklerik/urlshortener/internal/config"
	"github.com/Pklerik/urlshortener/internal/logger"
	"github.com/Pklerik/urlshortener/internal/middleware"
	"github.com/Pklerik/urlshortener/internal/repository"
	"github.com/Pklerik/urlshortener/internal/service"
	"github.com/go-chi/chi"
	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
)

func handler(parsedArgs *config.StartupFlags) http.Handler {

	linksRepo := repository.NewInMemoryLinksRepository()
	linksService := service.NewLinksService(linksRepo)
	linksHandler := NewLinkHandler(linksService, parsedArgs)
	r := chi.NewRouter()
	r.Route("/", func(r chi.Router) {
		r.Get("/{shortURL}", middleware.WithLogging(linksHandler.Get))
		r.Post("/", middleware.WithLogging(linksHandler.PostText))
		r.Route("/api", func(r chi.Router) {
			r.Post("/shorten", middleware.WithLogging(linksHandler.PostJSON))
		})
	})

	return r
}

func TestRegisterLinkHandler(t *testing.T) {
	type want struct {
		code                int
		response            string
		Headers             map[string][]string
		Location            string
		AcceptedContentType []string
	}
	testURL := []byte("http://ya.ru")
	redirectHost := "http://test_host:2345"
	testJSONReq := []byte("{\"url\":\"http://ya.ru\"}")
	testJSONResp := "\"result\""

	logger.Initialize("DEBUG")
	r := handler(&config.StartupFlags{
		BaseURL: redirectHost,
	})
	srv := httptest.NewServer(r)
	defer srv.Close()
	client := resty.New()
	client.SetRedirectPolicy(resty.NoRedirectPolicy())

	req := client.R()
	req.Method = http.MethodPost
	req.URL = srv.URL
	req.Body = testURL
	resp, err := req.Send()
	assert.NoError(t, err, "error making HTTP request")

	resBody := resp.Body()

	tests := []struct {
		name    string
		method  string
		headers map[string][]string
		body    *[]byte

		additionalURL string
		testError     string
		redirectHost  string
		want          want
	}{
		{name: "Post Created",
			method:       http.MethodPost,
			redirectHost: redirectHost,
			headers:      map[string][]string{"Content-Type": {"text/plain"}},
			body:         &testURL,
			want:         want{code: http.StatusCreated, response: redirectHost}},
		{name: "Wrong content",
			method:       http.MethodPost,
			redirectHost: redirectHost,
			headers:      map[string][]string{"Content-Type": {"application/json"}},
			body:         &testURL,
			want: want{
				code:     http.StatusBadRequest,
				response: "Wrong content type\n"}},
		{name: "Wrong Redirect",
			method:        http.MethodGet,
			redirectHost:  redirectHost,
			additionalURL: "/WERTADSD",
			want: want{
				code:     http.StatusBadRequest,
				response: "Unable to find long URL for short\n"}},
		{name: "Get Redirect",
			method:        http.MethodGet,
			redirectHost:  redirectHost,
			additionalURL: "/" + strings.Split(string(resBody), "/")[3], testError: "auto redirect is disabled",
			want: want{
				code:     http.StatusTemporaryRedirect,
				Headers:  map[string][]string{"Location": {string(testURL)}},
				response: ""}},
		{name: "Post Created JSON",
			method:        http.MethodPost,
			redirectHost:  redirectHost,
			additionalURL: "/api/shorten", headers: map[string][]string{"Content-Type": {"application/json"}},
			body: &testJSONReq,
			want: want{
				code:     http.StatusCreated,
				response: testJSONResp}},
		{name: "Wrong content JSON",
			method:        http.MethodPost,
			redirectHost:  redirectHost,
			additionalURL: "/api/shorten", headers: map[string][]string{"Content-Type": {"text/plain"}},
			body: &testJSONReq,
			want: want{
				code:     http.StatusBadRequest,
				response: "Wrong content type\n"}},
		{name: "Get Redirect GZIP",
			method:        http.MethodGet,
			redirectHost:  redirectHost,
			additionalURL: "/" + strings.Split(string(resBody), "/")[3], testError: "auto redirect is disabled",
			want: want{
				code:     http.StatusTemporaryRedirect,
				Headers:  map[string][]string{"Location": {string(testURL)}},
				response: ""}},
		{name: "Post Created JSON GZIP",
			method:        http.MethodPost,
			redirectHost:  redirectHost,
			additionalURL: "/api/shorten", headers: map[string][]string{"Content-Type": {"application/json"}, "Accept-Encoding": {"gzip"}},
			body: &testJSONReq,
			want: want{
				code:     http.StatusCreated,
				response: testJSONResp}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := client.R()
			req.Method = tt.method
			req.URL = srv.URL + tt.additionalURL
			if tt.body != nil {
				req.Body = *tt.body
			}
			req.SetHeaderMultiValues(tt.headers)

			resp, err := req.Send()

			if err != nil {
				assert.ErrorContains(t, err, tt.testError)
			}

			assert.Equal(t, tt.want.code, resp.StatusCode())
			respBody := string(resp.Body())
			assert.Contains(t, respBody, tt.want.response)
			if resp.Header().Get("Location") != "" {
				locations, ok := tt.want.Headers["Location"]
				if ok && len(locations) > 0 {
					assert.Contains(t, resp.Header().Values("Location"), locations[0])
				}
				assert.NotEmpty(t, tt.want.Headers["Location"], "For test not provided location")
			}

		})
	}
}
