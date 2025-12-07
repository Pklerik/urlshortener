// Package handler contains handling logic for all pages
package handler

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Pklerik/urlshortener/internal/config"
	"github.com/Pklerik/urlshortener/internal/logger"
	"github.com/Pklerik/urlshortener/internal/model"
	"github.com/Pklerik/urlshortener/internal/repository"
	mock_repository "github.com/Pklerik/urlshortener/internal/repository/mocks"
	"github.com/Pklerik/urlshortener/internal/service"
	"github.com/Pklerik/urlshortener/internal/service/links"
	"github.com/go-chi/chi"
	"github.com/golang/mock/gomock"
)

var (
	baseConfig = &config.StartupFlags{LocalStorage: "../../local_storage.json", LogLevel: "DEBUG", SecretKey: "fH72anZI1e6YFLN+Psh6Dv308js8Ul+q3mfPe8E36Qs="}
)

func init() {
	logger.Initialize(baseConfig.GetLogLevel())
}

func TestLinkHandle_Get(t *testing.T) {
	ctrl := gomock.NewController(t)
	r := mock_repository.NewMockLinksRepository(ctrl)

	defer ctrl.Finish()
	r.EXPECT().FindShort(gomock.Any(), "398f0ca4").Return(model.LinkData{UUID: "123", ShortURL: "398f0ca4", LongURL: "http://ya.ru"}, nil).AnyTimes()
	r.EXPECT().FindShort(gomock.Any(), gomock.Any()).Return(model.LinkData{UUID: "", ShortURL: "", LongURL: ""}, repository.ErrNotFoundLink).AnyTimes()
	type fields struct {
		Args config.StartupFlagsParser
	}
	type args struct {
		method   string
		target   string
		shortURL string
		body     io.Reader
	}

	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{name: "Get ya.ru",
			fields: fields{
				Args: &config.StartupFlags{}},
			args: args{
				method:   "GET",
				target:   "/",
				shortURL: "398f0ca4",
				body:     nil,
			}},
		{name: "empty Get",
			fields: fields{
				Args: &config.StartupFlags{}},
			args: args{
				method: "GET",
				target: "/",
				body:   nil,
			}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ls := links.NewLinksService(r, baseConfig.GetSecretKey())
			lh := NewLinkHandler(
				links.NewLinksService(r, baseConfig.GetSecretKey()),
				NewAuthenticationHandler(ls),
				tt.fields.Args,
			)
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("shortURL", tt.args.shortURL)
			req := httptest.NewRequest(tt.args.method, tt.args.target+tt.args.shortURL, tt.args.body)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
			lh.Get(httptest.NewRecorder(), req)
		})
	}
}

func TestLinkHandle_PostText(t *testing.T) {
	ctrl := gomock.NewController(t)
	r := mock_repository.NewMockLinksRepository(ctrl)

	defer ctrl.Finish()
	type fields struct {
		linkService service.LinkServicer
		Args        config.StartupFlagsParser
	}
	type args struct {
		w http.ResponseWriter
		r *http.Request
	}

	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{name: "base POST TEXT",
			fields: fields{
				linkService: links.NewLinksService(r, baseConfig.GetSecretKey()),
				Args:        &config.StartupFlags{}},
			args: args{
				w: httptest.NewRecorder(),
				r: httptest.NewRequest("POST", "/", bytes.NewBuffer([]byte("/")))}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lh := NewLinkHandler(
				tt.fields.linkService,
				NewAuthenticationHandler(tt.fields.linkService),
				tt.fields.Args,
			)
			lh.PostText(tt.args.w, tt.args.r)
		})
	}
}

func TestLinkHandle_PostJson(t *testing.T) {
	ctrl := gomock.NewController(t)
	r := mock_repository.NewMockLinksRepository(ctrl)
	type fields struct {
		linkService service.LinkServicer
		Args        config.StartupFlagsParser
	}
	type args struct {
		w http.ResponseWriter
		r *http.Request
	}

	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{name: "base POST JSON",
			fields: fields{
				linkService: links.NewLinksService(r, baseConfig.GetSecretKey()),
				Args:        &config.StartupFlags{}},
			args: args{
				w: httptest.NewRecorder(),
				r: httptest.NewRequest("POST", "/api/shorten", bytes.NewBuffer([]byte(`{"url": "/"}`)))}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lh := NewLinkHandler(
				tt.fields.linkService,
				NewAuthenticationHandler(tt.fields.linkService),
				tt.fields.Args,
			)
			lh.PostText(tt.args.w, tt.args.r)
		})
	}
}

func TestLinkHandle_PingDB(t *testing.T) {
	ctrl := gomock.NewController(t)
	r := mock_repository.NewMockLinksRepository(ctrl)

	defer ctrl.Finish()
	r.EXPECT().PingDB(gomock.Any()).Return(nil).AnyTimes()
	type fields struct {
		linkService service.LinkServicer
		Args        config.StartupFlagsParser
	}
	type args struct {
		w http.ResponseWriter
		r *http.Request
	}

	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{name: "base PING DB",
			fields: fields{
				linkService: links.NewLinksService(r, baseConfig.GetSecretKey()),
				Args:        baseConfig},
			args: args{
				w: httptest.NewRecorder(),
				r: httptest.NewRequest("GET", "/ping", nil)}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lh := NewLinkHandler(
				tt.fields.linkService,
				NewAuthenticationHandler(tt.fields.linkService),
				tt.fields.Args,
			)
			lh.PingDB(tt.args.w, tt.args.r)
		})
	}
}

func TestLinkHandle_PostBatchJSON(t *testing.T) {
	ctrl := gomock.NewController(t)
	r := mock_repository.NewMockLinksRepository(ctrl)

	defer ctrl.Finish()
	r.EXPECT().SetLinks(gomock.Any(), gomock.Any()).Return([]model.LinkData{{}, {}, {}}, nil).AnyTimes()

	type fields struct {
		linkService service.LinkServicer
		Args        config.StartupFlagsParser
	}
	type args struct {
		w http.ResponseWriter
		r *http.Request
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{name: "base POST JSON",
			fields: fields{
				linkService: links.NewLinksService(r, baseConfig.GetSecretKey()),
				Args:        &config.StartupFlags{}},
			args: args{
				w: httptest.NewRecorder(),
				r: httptest.NewRequest("POST", "/api/shorten/batch", bytes.NewBuffer([]byte(`[
    {
        "correlation_id": "req_id",
        "original_url": "http://ya.ru"
    },
    {
        "correlation_id": "req_id",
        "original_url": "http://yandex.ru"
    },
    {
        "correlation_id": "req_id",
        "original_url": "http://YAyandex.ru"
    }
]`)))}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.args.r.Header.Set("Content-Type", "application/json")
			lh := &LinkHandle{
				service: tt.fields.linkService,
				ah:      NewAuthenticationHandler(tt.fields.linkService),
				Args:    tt.fields.Args,
			}
			lh.PostBatchJSON(tt.args.w, tt.args.r)
		})
	}
}

// ExampleGet demonstrates how to use Get method of LinkHandler.
func ExampleGet() {
	ctrl := gomock.NewController(nil)
	r := mock_repository.NewMockLinksRepository(ctrl)

	defer ctrl.Finish()
	r.EXPECT().FindShort(gomock.Any(), "398f0ca4").Return(model.LinkData{UUID: "123", ShortURL: "398f0ca4", LongURL: "http://ya.ru"}, nil).AnyTimes()

	ls := links.NewLinksService(r, baseConfig.GetSecretKey())
	lh := NewLinkHandler(
		links.NewLinksService(r, baseConfig.GetSecretKey()),
		NewAuthenticationHandler(ls),
		baseConfig,
	)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("shortURL", "398f0ca4")
	req := httptest.NewRequest("GET", "/398f0ca4", nil)
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	lh.Get(httptest.NewRecorder(), req)
	// Output:
}
