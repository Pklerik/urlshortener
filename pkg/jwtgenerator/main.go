// Package jwtgenerator provide JWT token assembly and decryption funcs.
package jwtgenerator

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Pklerik/urlshortener/internal/logger"
	"github.com/golang-jwt/jwt/v4"
	"github.com/samborkent/uuidv7"
	"google.golang.org/grpc/metadata"
)

var (
	// ErrTokenParsing - error during token parsing.
	ErrTokenParsing = errors.New("error during token parsing")
	// ErrTokenValidation - error during token validation.
	ErrTokenValidation = errors.New("error token validation")
)

// ErrSigningMethodNotHMAC - error signing method not hmac.
type ErrSigningMethodNotHMAC struct {
	headerAlg string
}

func (e ErrSigningMethodNotHMAC) Error() string {
	return fmt.Sprint("error signing method not hmac: method:", e.headerAlg)
}

// Claims — структура утверждений, которая включает стандартные утверждения и
// одно пользовательское UserID.
type Claims struct {
	jwt.RegisteredClaims
	UserID uuidv7.UUID
}

// TokenExp time for token validation.
const TokenExp = time.Hour * 3

// BuildJWTString creates token and return it in string.
func BuildJWTString(userID uuidv7.UUID, secretKey string) (string, error) {
	// crate new token with algorithm fore sine  HS256  and some claims — Claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			// Crate Expire time for token.
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(TokenExp)),
		},
		// Payload
		UserID: userID,
	})

	// create token string
	tokenString, err := token.SignedString([]byte(secretKey))
	if err != nil {
		return "", fmt.Errorf("BuildJWTString: %w", err)
	}

	// возвращаем строку токена
	return tokenString, nil
}

// GetUserID provide UserId and error for secretKey and jwtToken.
func GetUserID(secretKey string, jwtToken string) (uuidv7.UUID, error) {
	claims := &Claims{}

	token, err := jwt.ParseWithClaims(jwtToken, claims, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrSigningMethodNotHMAC{headerAlg: fmt.Sprintf("%s", t.Header["alg"])}
		}

		return []byte(secretKey), nil
	})
	if err != nil {
		return uuidv7.New(), ErrTokenParsing
	}

	if !token.Valid {
		logger.Sugar.Error(ErrTokenValidation)
		return uuidv7.New(), ErrTokenValidation
	}

	return claims.UserID, nil
}

// ParseTokenFromCtxMetadata - proide string token from ctx metadata gRPC for first item in headerName.
func ParseTokenFromCtxMetadata(ctx context.Context, headerName string) (string, bool) {
	var (
		jwtToken string
		isValid  bool
	)

	md, ok := metadata.FromIncomingContext(ctx)
	if ok {
		values := md.Get(headerName)
		if len(values) > 0 {
			// ключ содержит слайс строк, получаем первую строку
			jwtToken = values[0]
			isValid = true
		}
	}

	return jwtToken, isValid
}
