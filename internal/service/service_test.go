// Package service provide all business logic for applications.
package service

import (
	"context"
	"reflect"
	"testing"

	"github.com/Pklerik/urlshortener/internal/logger"
	"github.com/Pklerik/urlshortener/internal/model"
	"github.com/Pklerik/urlshortener/internal/repository"
)

func TestBaseLinkService_RegisterLink(t *testing.T) {
	logger.Initialize("DEBUG")
	type fields struct {
		linksRepo repository.LinksStorager
	}
	type args struct {
		ctx     context.Context
		longURL string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    model.LinkData
		wantErr bool
	}{
		{name: "Base", fields: fields{linksRepo: repository.NewInMemoryLinksRepository()}, args: args{ctx: context.Background(), longURL: "http://ya.ru"}, want: model.LinkData{ShortURL: "398f0ca4", LongURL: "http://ya.ru"}, wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ls := NewLinksService(tt.fields.linksRepo)
			got, err := ls.RegisterLink(tt.args.ctx, tt.args.longURL)
			if (err != nil) != tt.wantErr {
				t.Errorf("BaseLinkService.RegisterLink() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("BaseLinkService.RegisterLink() = %v, want %v", got, tt.want)
			}
		})
	}
}
