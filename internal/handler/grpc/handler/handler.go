package handler

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	pb "github.com/Pklerik/urlshortener/api/proto"
	"github.com/Pklerik/urlshortener/internal/logger"
	"github.com/Pklerik/urlshortener/internal/model"
	"github.com/Pklerik/urlshortener/internal/repository"
	"github.com/Pklerik/urlshortener/internal/service"
	"github.com/Pklerik/urlshortener/pkg/jwtgenerator"
	"github.com/gogo/protobuf/proto"
	"github.com/samborkent/uuidv7"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/emptypb"
)

var (
	// ErrUnauthorizedUser - error for unauthorized user.
	ErrUnauthorizedUser = errors.New("unauthorized user")
)

// UsersLinksHandler поддерживает все необходимые методы сервера.
type UsersLinksHandler struct {
	// нужно встраивать тип pb.Unimplemented<TypeName>
	// для совместимости с будущими версиями
	pb.UnimplementedShortenerServiceServer

	// service предоставляет бизнес-логику для обработки запросов.
	service service.LinkServicer
}

// Проверка того, что UsersServer реализует интерфейс pb.ShortenerServiceServer.
var _ pb.ShortenerServiceServer = (*UsersLinksHandler)(nil)

func NewUsersLinksHandler(svc service.LinkServicer) http.Handler {
	ulh := &UsersLinksHandler{service: svc}
	s := grpc.NewServer(grpc.UnaryInterceptor(ulh.AuthHandle))
	pb.RegisterShortenerServiceServer(s, &UsersLinksHandler{service: svc})
	return s
}

func (us *UsersLinksHandler) AuthHandle(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
	err := us.authUser(ctx)
	if err != nil {
		return nil, fmt.Errorf("auth error: %w", err)
	}
	resp, err := handler(ctx, req)
	if err != nil {
		logger.Sugar.Errorf("AuthHandle %s,%v", info.FullMethod, err)
	} else {
		logger.Sugar.Infof("Auth User %s SUCCESS", info.FullMethod)
	}
	return resp, err

}

func (us *UsersLinksHandler) authUser(ctx context.Context) error {
	// Получаем ID пользователя из контекста
	var (
		jwtToken string
		err      error
	)

	md, ok := metadata.FromIncomingContext(ctx)
	if ok {
		values := md.Get("authorization")
		if len(values) > 0 {
			// ключ содержит слайс строк, получаем первую строку
			jwtToken = values[0]
		}
	}
	if jwtToken == "" {
		secret, ok := us.service.GetSecret("SECRET_KEY")
		if !ok {
			return service.ErrEmptySecret
		}
		jwtToken, err = jwtgenerator.BuildJWTString(uuidv7.New(), secret.(string))
		if err != nil {
			return fmt.Errorf("unable to getUserID: %w", err)
		}
	}

	headerMd := metadata.Pairs(
		"authorization", jwtToken,
	)
	if err := grpc.SendHeader(ctx, headerMd); err != nil {
		logger.Sugar.Warnf("could not send header: %v", err)
		// Handle error appropriately
	}

	return nil
}

func (us *UsersLinksHandler) GetUserID(ctx context.Context) (model.UserID, error) {
	// Получаем ID пользователя из контекста
	var (
		jwtToken string
		userID   model.UserID
		err      error
	)

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return userID, ErrUnauthorizedUser

	}

	values := md.Get("authorization")
	if len(values) == 0 {
		logger.Sugar.Warnf("unable to authorized user")
		return userID, ErrUnauthorizedUser
	}
	// ключ содержит слайс строк, получаем первую строку
	jwtToken = values[0]
	secret, ok := us.service.GetSecret("SECRET_KEY")
	if !ok {
		return userID, service.ErrEmptySecret
	}

	userIDUUID, err := jwtgenerator.GetUserID(secret.(string), jwtToken)
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
		return nil, err
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

func (us *UsersLinksHandler) ListUserURLs(ctx context.Context, _ *emptypb.Empty) (*pb.UserURLsResponse, error) {
	// Реализация метода ListUserURLs
	userID, err := us.GetUserID(ctx)
	if err != nil {
		return &pb.UserURLsResponse{}, fmt.Errorf("ListUserURLs: %w", err)
	}
	links, err := us.service.ProvideUserLinks(ctx, userID)
	if err != nil {
		return nil, err
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
