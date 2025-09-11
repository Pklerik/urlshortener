// Package handler contains handling logic for all pages
package handler

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Pklerik/urlshortener/internal/config"
	"github.com/Pklerik/urlshortener/internal/logger"
	"github.com/Pklerik/urlshortener/internal/repository"
	"github.com/Pklerik/urlshortener/internal/service"
)

var (
	baseConfig = &config.StartupFlags{LocalStorage: "local_storage.json", LogLevel: "DEBUG"}
)

func init() {
	logger.Initialize(baseConfig.GetLogLevel())
}

func TestLinkHandle_Get(t *testing.T) {
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
		{name: "base Get",
			fields: fields{
				linkService: service.NewLinksService(repository.NewLocalMemoryLinksRepository(baseConfig.GetLocalStorage())),
				Args:        &config.StartupFlags{}},
			args: args{
				w: httptest.NewRecorder(),
				r: httptest.NewRequest("GET", "/", nil)}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lh := NewLinkHandler(
				tt.fields.linkService,
				tt.fields.Args,
			)
			lh.Get(tt.args.w, tt.args.r)
		})
	}
}

func TestLinkHandle_PostText(t *testing.T) {
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
				linkService: service.NewLinksService(repository.NewLocalMemoryLinksRepository(baseConfig.GetLocalStorage())),
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
				linkService: service.NewLinksService(repository.NewLocalMemoryLinksRepository(baseConfig.GetLocalStorage())),
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
				linkService: service.NewLinksService(repository.NewLocalMemoryLinksRepository(baseConfig.LocalStorage)),
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
