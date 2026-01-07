package handler

import (
	"context"
	"crypto/rand"
	"fmt"
	"log"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"syscall"
	"testing"

	pb "github.com/Pklerik/urlshortener/api/proto"
	"github.com/Pklerik/urlshortener/internal/logger"
	"github.com/Pklerik/urlshortener/internal/repository/inmemory"
	"github.com/Pklerik/urlshortener/internal/service/links"
	"github.com/Pklerik/urlshortener/pkg/jwtgenerator"
	"github.com/samborkent/uuidv7"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/proto"
)

func Test_Handler(t *testing.T) {
	// initialize logger used by service
	_ = logger.Initialize("debug")
	port := ":50505"
	ctx := context.Background()

	cancel, err := StartSever(ctx, port)
	defer cancel()

	// Устанавливаем соединение с сервером
	addr := "127.0.0.1" + port
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		slog.Error("ошибка при установлении соединения с сервером", "error", err)
		t.Error("Error running grpc client")
	}
	defer conn.Close()
	c := pb.NewShortenerServiceClient(conn)

	// функция, в которой будем отправлять сообщения
	if err := SendUsersRequests(ctx, c); err != nil {
		slog.Error("ошибка при работе клиента", "error", err)
		t.Error("Error running grpc client")
	}

	cancel()
}

func SendUsersRequests(ctx context.Context, c pb.ShortenerServiceClient) error {
	b := make([]byte, 20)
	_, _ = rand.Read(b)

	link := "https://example.com/some/very/long/url" + fmt.Sprintf("%x", b)

	// build jwt token and attach to metadata
	uid := uuidv7.New()
	token, err := jwtgenerator.BuildJWTString(uid, "secret_key")
	if err != nil {
		return err
	}

	header := metadata.New(map[string]string{"authorization": token})
	ctx = metadata.NewOutgoingContext(ctx, header)
	// Отправляем запрос на сокращение URL
	shortenResp, err := c.ShortenURL(ctx, pb.URLShortenRequest_builder{Url: proto.String(link)}.Build())
	if err != nil {
		return err
	}
	slog.Info("Shortened URL", "short_url", shortenResp.GetResult())

	expandResp, err := c.ExpandURL(ctx, pb.URLExpandRequest_builder{Id: proto.String(shortenResp.GetResult())}.Build())
	if err != nil {
		return err
	}
	slog.Info("Expand URL", "short_url", expandResp.GetResult())
	return nil
}

func StartSever(ctx context.Context, port string) (context.CancelFunc, error) {
	ctx, cancel := context.WithCancel(ctx)

	go func() {
		c := make(chan os.Signal, 1) // we need to reserve to buffer size 1, so the notifier are not blocked
		signal.Notify(c, os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGINT)

		<-c
		cancel()
	}()

	linksRepo := inmemory.NewInMemoryLinksRepository()

	linksService := links.NewLinksService(linksRepo, "secret_key")
	uhlService := NewUsersLinksHandler(ctx, linksService).Register()

	lis, err := net.Listen("tcp", "127.0.0.1"+port)
	if err != nil {
		log.Println("failed to listen:", err)
		cancel()
		return cancel, err
	}

	g, gCtx := errgroup.WithContext(ctx)
	g.Go(func() error {
		log.Println("Starting gRPC server")
		return uhlService.Serve(lis)
	})
	g.Go(func() error {
		<-gCtx.Done()
		log.Println("Stopped serving new connections.")
		return lis.Close()
	})
	return cancel, nil

}
