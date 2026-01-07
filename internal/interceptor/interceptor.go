// Package interceptor provide intercepting func for gRPC.
package interceptor

import (
	"context"
	"errors"
	"fmt"

	"github.com/Pklerik/urlshortener/internal/logger"
	"github.com/Pklerik/urlshortener/pkg/jwtgenerator"
	"github.com/samborkent/uuidv7"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

var (
	ErrEmptyMetadata = errors.New("empty metadata")
	ErrSendingHeader = errors.New("unable to send header")
)

// AuthUnaryServerInterceptor - provide interceptor for authorization on gRPC requests.
func AuthUnaryServerInterceptor(secret string) func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
	fn := func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		ctx, err := authUser(ctx, secret)
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

	return fn
}

func authUser(ctx context.Context, secret string) (context.Context, error) {
	var err error
	// Получаем ID пользователя из контекста
	headerName := "authorization"

	jwtToken, ok := jwtgenerator.ParseTokenFromCtxMetadata(ctx, headerName)
	if !ok {
		return ctx, fmt.Errorf("authUser: %w", err)
	}

	if jwtToken == "" {
		jwtToken, err = jwtgenerator.BuildJWTString(uuidv7.New(), secret)
		if err != nil {
			return ctx, fmt.Errorf("unable to getUserID: %w", err)
		}
	}

	newCtx, err := updateAuthorizationToken(ctx, jwtToken, headerName)
	if err != nil {
		return ctx, fmt.Errorf("authUser: %w", err)
	}

	return newCtx, nil
}

func updateAuthorizationToken(ctx context.Context, jwtToken, headerName string) (context.Context, error) {
	headerMd := metadata.Pairs(
		headerName, jwtToken,
	)
	if err := grpc.SendHeader(ctx, headerMd); err != nil {
		logger.Sugar.Warnf("could not send header: %v", err)
		return ctx, ErrSendingHeader
	}

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ctx, ErrEmptyMetadata
	}
	md.Append(headerName, jwtToken)
	return metadata.NewIncomingContext(ctx, md), nil
}
