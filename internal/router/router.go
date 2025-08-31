// Package router provides functionality for setups and startup of server.
package router

import (
	"net/http"

	//nolint

	"github.com/Pklerik/urlshortener/internal/config"
	"github.com/Pklerik/urlshortener/internal/handler"
	"github.com/Pklerik/urlshortener/internal/middleware"
	"github.com/Pklerik/urlshortener/internal/repository"
	"github.com/Pklerik/urlshortener/internal/service"
	"github.com/go-chi/chi"
)

// ConfigureRouter starts server with base configuration.
func ConfigureRouter(parsedFlags config.StartupFlagsParser) http.Handler {
	linksRepo := repository.NewInMemoryLinksRepository()
	linksService := service.NewLinksService(linksRepo)
	linksHandler := handler.NewLinkHandler(linksService, parsedFlags)
	r := chi.NewRouter()
	r.Route("/", func(r chi.Router) {
		r.Get("/{shortURL}", middleware.WithLogging(middleware.GZIPMiddleware(linksHandler.Get)))
		r.Post("/", middleware.WithLogging(middleware.GZIPMiddleware(linksHandler.PostText)))
		r.Route("/api", func(r chi.Router) {
			r.Post("/shorten", middleware.WithLogging(middleware.GZIPMiddleware(linksHandler.PostJSON)))
		})
	})

	return r
}
