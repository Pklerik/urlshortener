// Package handler provide gRPC handler service realization.
package handler

import (
	"context"
	"errors"
	"fmt"

	pb "github.com/Pklerik/urlshortener/api/proto"
	"github.com/Pklerik/urlshortener/internal/logger"
	"github.com/Pklerik/urlshortener/internal/model"
	"github.com/Pklerik/urlshortener/internal/repository"
	"github.com/Pklerik/urlshortener/internal/service"
	"github.com/Pklerik/urlshortener/pkg/jwtgenerator"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/emptypb"
)

var (
	// ErrUnauthorizedUser - error for unauthorized user.
	ErrUnauthorizedUser = errors.New("unauthorized user")
	// ErrEmptyToken unable to parse token.
	ErrEmptyToken = errors.New("unable to parse token")
)

// UsersLinksHandler поддерживает все необходимые методы сервера.
type UsersLinksHandler struct {
	// нужно встраивать тип pb.Unimplemented<TypeName>
	// для совместимости с будущими версиями
	pb.UnimplementedShortenerServiceServer

	// service предоставляет бизнес-логику для обработки запросов.
	service service.LinkServicer
}

// Проверка того, что UsersLinksHandler реализует интерфейс pb.ShortenerServiceServer.
var _ pb.ShortenerServiceServer = (*UsersLinksHandler)(nil)

// NewUsersLinksHandler - provide gRPC Handlers for Links Service.
func NewUsersLinksHandler(_ context.Context, svc service.LinkServicer) *UsersLinksHandler {
	return &UsersLinksHandler{service: svc}

}

func (ulh *UsersLinksHandler) Register(interceptors ...grpc.UnaryServerInterceptor) *grpc.Server {
	s := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			interceptors...))
	pb.RegisterShortenerServiceServer(s, ulh)

	return s
}

// GetUserID return userID provided from JWT token form context.
func (us *UsersLinksHandler) GetUserID(ctx context.Context) (model.UserID, error) {
	// Получаем ID пользователя из контекста
	var (
		userID model.UserID
		err    error
	)

	jwtToken, ok := jwtgenerator.ParseTokenFromCtxMetadata(ctx, "authorization")
	if !ok {
		return userID, ErrEmptyToken
	}

	secret, ok := us.service.GetSecret("SECRET_KEY")
	if !ok {
		return userID, service.ErrEmptySecret
	}

	userIDUUID, err := jwtgenerator.GetUserID(secret, jwtToken)
	if err != nil {
		return userID, fmt.Errorf("unexpected JWT parsing error: %w", err)
	}

	return model.UserID(userIDUUID.String()), nil
}

// Здесь будут реализованы методы сервера

// ExpandURL реализует метод получения URL по короткой ссылке.
func (us *UsersLinksHandler) ExpandURL(ctx context.Context, req *pb.URLExpandRequest) (*pb.URLExpandResponse, error) {
	link, err := us.service.GetShort(ctx, req.GetId())
	if err != nil {
		return nil, fmt.Errorf("ExpandURL: %w", err)
	}

	response := pb.URLExpandResponse_builder{
		Result: proto.String(link.LongURL),
	}.Build()

	return response, nil
}

// ShortenURL реализует метод сокращения URL.
func (us *UsersLinksHandler) ShortenURL(ctx context.Context, req *pb.URLShortenRequest) (*pb.URLShortenResponse, error) {
	// Реализация метода ShortenURL
	userID, err := us.GetUserID(ctx)
	if err != nil {
		return &pb.URLShortenResponse{}, fmt.Errorf("ShortenURL: %w", err)
	}

	links, err := us.service.RegisterLinks(ctx, []string{req.GetUrl()}, userID)
	if err != nil && !errors.Is(err, repository.ErrExistingLink) {
		logger.Sugar.Errorf("grpc ShortenURL error: %v", err)
		return nil, fmt.Errorf("ShortenURL: %w", err)
	}

	if len(links) == 0 {
		return &pb.URLShortenResponse{}, nil
	}

	response := pb.URLShortenResponse_builder{
		Result: proto.String(links[0].ShortURL),
	}.Build()

	return response, nil
}

// ListUserURLs provide service statistics.
func (us *UsersLinksHandler) ListUserURLs(ctx context.Context, _ *emptypb.Empty) (*pb.UserURLsResponse, error) {
	// Реализация метода ListUserURLs
	userID, err := us.GetUserID(ctx)
	if err != nil {
		return &pb.UserURLsResponse{}, fmt.Errorf("ListUserURLs: %w", err)
	}

	links, err := us.service.ProvideUserLinks(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("ListUserURLs: %w", err)
	}

	response := pb.UserURLsResponse_builder{
		Url: make([]*pb.URLData, len(links)),
	}.Build()

	urls := make([]*pb.URLData, 0, len(links))
	for _, link := range links {
		url := pb.URLData_builder{
			ShortUrl:    proto.String(link.ShortURL),
			OriginalUrl: proto.String(link.LongURL),
		}.Build()
		urls = append(urls, url)
	}

	response.SetUrl(urls)

	return response, nil
}
