package jwtgenerator

import (
	"errors"
	"fmt"
	"time"

	"github.com/Pklerik/urlshortener/internal/logger"
	"github.com/golang-jwt/jwt/v4"
)

var (
	// ErrTokenParsing - error during token parsing.
	ErrTokenParsing = errors.New("error during token parsing")
	// ErrTokenValidation - error during token validation.
	ErrTokenValidation = errors.New("error token validation")
)

// Claims — структура утверждений, которая включает стандартные утверждения и
// одно пользовательское UserID
type Claims struct {
	jwt.RegisteredClaims
	UserID int
}

const TOKEN_EXP = time.Hour * 3

// BuildJWTString создаёт токен и возвращает его в виде строки.
func BuildJWTString(userID int, secretKey string) (string, error) {
	// создаём новый токен с алгоритмом подписи HS256 и утверждениями — Claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			// когда создан токен
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(TOKEN_EXP)),
		},
		// собственное утверждение
		UserID: userID,
	})

	// создаём строку токена
	tokenString, err := token.SignedString([]byte(secretKey))
	if err != nil {
		return "", err
	}

	// возвращаем строку токена
	return tokenString, nil
}

func GetUserID(secretKey string, jwtToken string) (int, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(jwtToken, claims, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(secretKey), nil
	})
	if err != nil {
		return -1, ErrTokenParsing
	}
	if !token.Valid {
		logger.Sugar.Error(ErrTokenValidation)
		return -1, ErrTokenValidation
	}
	return claims.UserID, nil
}
