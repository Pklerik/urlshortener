// Package router provides functionality for setups and startup of server.
package router

import (
	"net/http"

	//nolint

	"github.com/Pklerik/urlshortener/internal/handler"
	"github.com/Pklerik/urlshortener/internal/repository"
	"github.com/Pklerik/urlshortener/internal/service"
	"github.com/go-chi/chi"
)

// ConfigureRouter starts server with base configuration.
func ConfigureRouter() http.Handler {
	linksRepo := repository.NewInMemoryLinksRepository()
	linksService := service.NewLinksService(linksRepo)
	linksHandler := handler.NewLinkHandler(linksService)
	r := chi.NewRouter()
	r.Route("/", func(r chi.Router) {
		r.Get("/{shortURL}", linksHandler.GetRegisterLinkHandler)
		r.Post("/", linksHandler.PostRegisterLinkHandler)
	})

	return r
}
