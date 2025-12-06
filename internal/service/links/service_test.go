// Package links provide all business logic for links shortening app.
package links

import (
	"context"
	"reflect"
	"testing"

	"github.com/Pklerik/urlshortener/internal/logger"
	"github.com/Pklerik/urlshortener/internal/model"
	"github.com/Pklerik/urlshortener/internal/repository"
	"github.com/Pklerik/urlshortener/internal/repository/inmemory"
	"github.com/Pklerik/urlshortener/internal/repository/mocks"
	"github.com/golang/mock/gomock"
)

func TestBaseLinkService_RegisterLinks(t *testing.T) {
	logger.Initialize("DEBUG")
	type fields struct {
		linksRepo repository.LinksRepository
	}
	type args struct {
		ctx      context.Context
		longURLs []string
	}
	tests := []struct {
		fields  fields
		name    string
		want    string
		args    args
		wantErr bool
	}{
		{name: "Base", fields: fields{linksRepo: inmemory.NewInMemoryLinksRepository()}, args: args{ctx: context.Background(), longURLs: []string{"http://ya.ru"}}, want: "398f0ca4", wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ls := NewLinksService(tt.fields.linksRepo, "fH72anZI1e6YFLN+Psh6Dv308js8Ul+q3mfPe8E36Qs=")
			gots, err := ls.RegisterLinks(tt.args.ctx, tt.args.longURLs, "0199996a-fd98-780c-b5aa-1aef966fb36e0")
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

func BenchmarkBaseLinkService_RegisterLinks_Single(b *testing.B) {
	logger.Initialize("ERROR")
	ctrl := gomock.NewController(b)
	mockRepo := mocks.NewMockLinksRepository(ctrl)
	mockRepo.EXPECT().CreateUser(gomock.Any(), gomock.Any()).Return(model.User{ID: "1"}, nil).AnyTimes()
	mockRepo.EXPECT().SetLinks(gomock.Any(), gomock.Any()).Return([]model.LinkData{}, nil).AnyTimes()
	mockRepo.EXPECT().FindShort(gomock.Any(), "398f0ca4").Return(model.LinkData{UUID: "0199996a-fd98-780c-b5aa-1aef966fb36e0", ShortURL: "398f0ca4", LongURL: "http://ya.ru"}, nil).AnyTimes()
	ls := NewLinksService(mockRepo, "fH72anZI1e6YFLN+Psh6Dv308js8Ul+q3mfPe8E36Qs=")
	ctx := context.Background()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, err := ls.RegisterLinks(ctx, []string{"http://ya.ru"}, "0199996a-fd98-780c-b5aa-1aef966fb36e0"); err != nil {
			b.Fatalf("RegisterLinks error: %v", err)
		}
	}
}

func BenchmarkBaseLinkService_RegisterLinks_Parallel(b *testing.B) {
	logger.Initialize("ERROR")
	ctrl := gomock.NewController(b)
	mockRepo := mocks.NewMockLinksRepository(ctrl)
	mockRepo.EXPECT().CreateUser(gomock.Any(), gomock.Any()).Return(model.User{ID: "1"}, nil).AnyTimes()
	mockRepo.EXPECT().SetLinks(gomock.Any(), gomock.Any()).Return([]model.LinkData{}, nil).AnyTimes()
	mockRepo.EXPECT().FindShort(gomock.Any(), "398f0ca4").Return(model.LinkData{UUID: "0199996a-fd98-780c-b5aa-1aef966fb36e0", ShortURL: "398f0ca4", LongURL: "http://ya.ru"}, nil).AnyTimes()
	ls := NewLinksService(mockRepo, "fH72anZI1e6YFLN+Psh6Dv308js8Ul+q3mfPe8E36Qs=")
	ctx := context.Background()
	b.ReportAllocs()
	b.SetParallelism(4)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			if _, err := ls.RegisterLinks(ctx, []string{"http://ya.ru"}, "0199996a-fd98-780c-b5aa-1aef966fb36e0"); err != nil {
				b.Fatalf("RegisterLinks error: %v", err)
			}
		}
	})
}
