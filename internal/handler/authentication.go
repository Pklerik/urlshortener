package handler

import (
	"errors"
	"net/http"

	"github.com/Pklerik/urlshortener/internal/logger"
	"github.com/Pklerik/urlshortener/internal/model"
	"github.com/Pklerik/urlshortener/internal/service"
	"github.com/Pklerik/urlshortener/pkg/jwtgenerator"
	"github.com/samborkent/uuidv7"
)

// IAuthentication provide user authentication middleware.
type IAuthentication interface {
	AuthUser(next http.Handler) http.Handler
	GetUserIDFromCookie(r *http.Request) (model.UserID, error)
}

var (
	// ErrUnauthorizedUser - error for unauthorized user.
	ErrUnauthorizedUser = errors.New("unauthorized user")
)

// AuthHandler provide user authentication middleware.
type AuthHandler struct {
	service service.LinkServicer
}

// NewAuthenticationHandler provide new instance of AuthHandler.
func NewAuthenticationHandler(service service.LinkServicer) *AuthHandler {
	return &AuthHandler{
		service: service,
	}
}

// AuthUser provide middleware for user authentication.
func (ah *AuthHandler) AuthUser(next http.Handler) http.Handler {
	cookieName := "auth_user"
	authFn := func(w http.ResponseWriter, r *http.Request) {
		cookieAuth, err := r.Cookie(cookieName)
		if err != nil && !errors.Is(err, http.ErrNoCookie) {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if errors.Is(err, http.ErrNoCookie) {
			secretKey, ok := ah.service.GetSecret("SECRET_KEY")
			if !ok {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			validJWT, err := jwtgenerator.BuildJWTString(
				uuidv7.New(),
				secretKey.(string),
			)
			if err != nil {
				http.Error(w, "Unable to generete JWT", http.StatusInternalServerError)
				logger.Sugar.Errorf("Unable to generete JWT: %w", err)

				return
			}

			cookieAuth = &http.Cookie{
				Name:  cookieName,
				Value: validJWT,
				Path:  "/",
			}
			r.AddCookie(cookieAuth)
		}

		http.SetCookie(w, cookieAuth)

		// передаём управление хендлеру
		next.ServeHTTP(w, r)
	}

	return http.HandlerFunc(authFn)
}

// GetUserIDFromCookie provide userId auth info.
func (ah *AuthHandler) GetUserIDFromCookie(r *http.Request) (model.UserID, error) {
	authCookie, err := r.Cookie("auth_user")
	if errors.Is(err, http.ErrNoCookie) {
		logger.Sugar.Infof(`Unable to find auth_user cookie: %d`, http.StatusUnauthorized)

		return model.UserID(uuidv7.New().String()), ErrUnauthorizedUser
	}

	if err != nil {
		logger.Sugar.Infof(`Unable to get cookie: status: %d`, http.StatusInternalServerError)
	}

	secretKey, ok := ah.service.GetSecret("SECRET_KEY")
	if !ok {
		return model.UserID(uuidv7.New().String()), ErrUnauthorizedUser
	}

	userID, err := jwtgenerator.GetUserID(secretKey.(string), authCookie.Value)
	if err != nil {
		logger.Sugar.Infof(`Unable to get UserID: status: %d`, http.StatusUnauthorized)

		return model.UserID(userID.String()), ErrUnauthorizedUser
	}

	return model.UserID(model.UUIDv7(userID.String())), nil
}
