package handler

import (
	"context"
	"crypto/rand"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"testing"

	pb "github.com/Pklerik/urlshortener/api/proto"
	"github.com/Pklerik/urlshortener/internal/repository/inmemory"
	"github.com/Pklerik/urlshortener/internal/service/links"
	"github.com/gogo/protobuf/proto"
	"golang.org/x/net/http2/h2c"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

func Test_Handler(t *testing.T) {
	port := ":50505"
	ctx := context.Background()

	var (
		cancel context.CancelFunc
		err    error
	)
	defer cancel()
	g, _ := errgroup.WithContext(ctx)
	g.Go(func() error {
		cancel, err = StartSever(ctx, port)
		return err
	},
	)

	// Устанавливаем соединение с сервером
	conn, err := grpc.NewClient(port, grpc.WithTransportCredentials(insecure.NewCredentials()))
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
}

func SendUsersRequests(ctx context.Context, c pb.ShortenerServiceClient) error {

	b := make([]byte, 20)
	_, _ = rand.Read(b)

	link := "https://example.com/some/very/long/url" + string(b)

	header := metadata.New(map[string]string{})
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
	ulh, err := NewUsersLinksHandler(ctx, linksService)
	if err != nil {
		return cancel, fmt.Errorf("Unexpected error: %w", err)
	}

	httpServer := &http.Server{
		Addr:    port,
		Handler: h2c.NewHandler(ulh, nil),
	}

	g, gCtx := errgroup.WithContext(ctx)
	g.Go(func() error {
		log.Println("Starting server")

		return httpServer.ListenAndServe()
	})
	g.Go(func() error {
		<-gCtx.Done()
		log.Println("Stopped serving new connections.")

		return httpServer.Shutdown(context.Background())
	})

	if err := g.Wait(); err != nil {
		log.Printf("exit reason: %s \n", err)
	}
	return cancel, nil

}
