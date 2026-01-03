package interceptor

import "net/http"

func NewAuthInterceptor() *AuthInterceptor {
	return &AuthInterceptor{}
}

// AuthInterceptor provide struct for intercepting requests.
type AuthInterceptor struct{}

func (i *AuthInterceptor) Intercept(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Здесь можно добавить логику аутентификации или авторизации
		// Например, проверка наличия и валидности токена в заголовках запроса

		// Если аутентификация прошла успешно, передаем управление следующему обработчику
		next.ServeHTTP(w, r)
	})
}
