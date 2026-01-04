// Package interceptor provide intercepting func for gRPC.
package interceptor

import (
	"context"
	"fmt"

	"github.com/Pklerik/urlshortener/internal/logger"
	"github.com/Pklerik/urlshortener/pkg/jwtgenerator"
	"github.com/samborkent/uuidv7"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
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
		jwtToken, err = jwtgenerator.BuildJWTString(uuidv7.New(), secret)
		if err != nil {
			return ctx, fmt.Errorf("unable to getUserID: %w", err)
		}
	}

	headerMd := metadata.Pairs(
		"authorization", jwtToken,
	)
	if err := grpc.SendHeader(ctx, headerMd); err != nil {
		logger.Sugar.Warnf("could not send header: %v", err)
		// Handle error appropriately
	}
	md.Append("authorization", jwtToken)
	newCtx := metadata.NewIncomingContext(ctx, md)

	return newCtx, nil
}
