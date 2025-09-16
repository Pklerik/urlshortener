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
	"github.com/go-chi/chi"
	"github.com/golang/mock/gomock"
)

var (
	baseConfig = &config.StartupFlags{LocalStorage: "../../local_storage.json", LogLevel: "DEBUG"}
)

func init() {
	logger.Initialize(baseConfig.GetLogLevel())
}

func TestLinkHandle_Get(t *testing.T) {
	ctrl := gomock.NewController(t)
	r := mock_repository.NewMockLinksStorager(ctrl)

	defer ctrl.Finish()
	r.EXPECT().FindShort(gomock.Any(), "398f0ca4").Return(model.LinkData{UUID: "123", ShortURL: "398f0ca4", LongURL: "http://ya.ru"}, nil).AnyTimes()
	r.EXPECT().FindShort(gomock.Any(), gomock.Any()).Return(model.LinkData{UUID: "", ShortURL: "", LongURL: ""}, repository.ErrNotFoundLink).AnyTimes()
	type fields struct {
		Args config.StartupFlagsParser
	}
	type args struct {
		method   string
		target   string
		shortURl string
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
				shortURl: "398f0ca4",
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
			lh := NewLinkHandler(
				service.NewLinksService(r),
				tt.fields.Args,
			)
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("shortURL", tt.args.shortURl)
			req := httptest.NewRequest(tt.args.method, tt.args.target+tt.args.shortURl, tt.args.body)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
			lh.Get(httptest.NewRecorder(), req)
		})
	}
}

func TestLinkHandle_PostText(t *testing.T) {
	ctrl := gomock.NewController(t)
	r := mock_repository.NewMockLinksStorager(ctrl)

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
				linkService: service.NewLinksService(r),
				Args:        &config.StartupFlags{}},
			args: args{
				w: httptest.NewRecorder(),
				r: httptest.NewRequest("POST", "/", bytes.NewBuffer([]byte("/")))}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lh := NewLinkHandler(
				tt.fields.linkService,
				tt.fields.Args,
			)
			lh.PostText(tt.args.w, tt.args.r)
		})
	}
}

func TestLinkHandle_PostJson(t *testing.T) {
	ctrl := gomock.NewController(t)
	r := mock_repository.NewMockLinksStorager(ctrl)
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
				linkService: service.NewLinksService(r),
				Args:        &config.StartupFlags{}},
			args: args{
				w: httptest.NewRecorder(),
				r: httptest.NewRequest("POST", "/api/shorten", bytes.NewBuffer([]byte(`{"url": "/"}`)))}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lh := NewLinkHandler(
				tt.fields.linkService,
				tt.fields.Args,
			)
			lh.PostText(tt.args.w, tt.args.r)
		})
	}
}

func TestLinkHandle_PingDB(t *testing.T) {
	ctrl := gomock.NewController(t)
	r := mock_repository.NewMockLinksStorager(ctrl)

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
				linkService: service.NewLinksService(r),
				Args:        baseConfig},
			args: args{
				w: httptest.NewRecorder(),
				r: httptest.NewRequest("GET", "/ping", nil)}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lh := NewLinkHandler(
				tt.fields.linkService,
				tt.fields.Args,
			)
			lh.PingDB(tt.args.w, tt.args.r)
		})
	}
}
