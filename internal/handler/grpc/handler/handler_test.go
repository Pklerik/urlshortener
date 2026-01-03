package handler

import (
	"context"
	"crypto/tls"
	"log/slog"
	"os"
	"testing"

	pb "github.com/Pklerik/urlshortener/api/proto"
	"github.com/gogo/protobuf/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
)

func Test_Handler(t *testing.T) {
	ctx := context.Background()

	// Load TLS certificate
	cert, err := tls.LoadX509KeyPair("../../../../cert/cert.pem", "../../../../cert/private.pem")
	if err != nil {
		slog.Error("ошибка при загрузке сертификата", "error", err)
		os.Exit(1)
	}

	tlsConfig := &tls.Config{
		Certificates:       []tls.Certificate{cert},
		InsecureSkipVerify: true, // Skip verification for self-signed certs
	}
	creds := credentials.NewTLS(tlsConfig)

	// Устанавливаем соединение с сервером
	conn, err := grpc.NewClient(`:8080`, grpc.WithTransportCredentials(creds))
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
	link := "https://example.com/some/very/long/url"

	header := metadata.New(map[string]string{"authorization": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3Njc0NzQxNjcsIlVzZXJJRCI6WzEsMTU1LDEzMyw2LDE3OCwxMzIsMTI2LDg2LDE5MCwyMzYsMTg4LDEzMyw1MiwyMTEsMTE4LDEwNl19.UzWIDhBB-rdREeF5pJacGG_Oh6nTafN9oqrOBtaRxv4"})

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
