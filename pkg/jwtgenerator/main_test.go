package jwtgenerator

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/samborkent/uuidv7"
)

func TestBuildJWTString(t *testing.T) {
	tests := []struct {
		name      string
		secretKey string
		userID    uuidv7.UUID
		wantErr   bool
	}{
		{
			name:      "valid token creation",
			userID:    uuidv7.New(),
			secretKey: "test-secret-key",
			wantErr:   false,
		},
		{
			name:      "empty secret key",
			userID:    uuidv7.New(),
			secretKey: "",
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokenString, err := BuildJWTString(tt.userID, tt.secretKey)

			if (err != nil) != tt.wantErr {
				t.Errorf("BuildJWTString() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tokenString == "" {
				t.Error("BuildJWTString() returned empty token string")
				return
			}

			// Verify token can be parsed and contains correct userID
			claims := &Claims{}
			token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
				return []byte(tt.secretKey), nil
			})

			if err != nil {
				t.Errorf("Failed to parse generated token: %v", err)
				return
			}

			if !token.Valid {
				t.Error("Generated token is invalid")
				return
			}

			if claims.UserID != tt.userID {
				t.Errorf("Token UserID = %v, want %v", claims.UserID, tt.userID)
			}

			if claims.ExpiresAt == nil {
				t.Error("Token ExpiresAt is nil")
				return
			}

			if claims.ExpiresAt.Before(time.Now()) {
				t.Error("Token is already expired")
			}
		})
	}
}
