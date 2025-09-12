// Package service provide all business logic for applications.
package service

import (
	"context"
	"reflect"
	"testing"

	"github.com/Pklerik/urlshortener/internal/logger"
	"github.com/Pklerik/urlshortener/internal/repository"
)

func TestBaseLinkService_RegisterLinks(t *testing.T) {
	logger.Initialize("DEBUG")
	type fields struct {
		linksRepo repository.LinksStorager
	}
	type args struct {
		ctx      context.Context
		longURLs []string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    string
		wantErr bool
	}{
		{name: "Base", fields: fields{linksRepo: repository.NewInMemoryLinksRepository()}, args: args{ctx: context.Background(), longURLs: []string{"http://ya.ru"}}, want: "398f0ca4", wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ls := NewLinksService(tt.fields.linksRepo)
			gots, err := ls.RegisterLinks(tt.args.ctx, tt.args.longURLs)
			if (err != nil) != tt.wantErr {
				t.Errorf("BaseLinkService.RegisterLinks() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			for _, got := range gots {
				if !reflect.DeepEqual(got.ShortURL, tt.want) {
					t.Errorf("BaseLinkService.RegisterLinks() = %v, want %v", got, tt.want)
				}
			}
		})
	}
}
