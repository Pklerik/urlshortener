// Package router provides functionality for setups and startup of server.
package router

import (
	"database/sql"
	"net/http"

	//nolint

	"github.com/Pklerik/urlshortener/internal/config"
	"github.com/Pklerik/urlshortener/internal/handler"
	"github.com/Pklerik/urlshortener/internal/internalmiddleware"
	"github.com/Pklerik/urlshortener/internal/repository"
	"github.com/Pklerik/urlshortener/internal/service"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

// ConfigureRouter starts server with base configuration.
func ConfigureRouter(parsedFlags config.StartupFlagsParser, db *sql.DB) http.Handler {
	var linksRepo repository.LinksStorager
	switch {
	case db != nil:
		linksRepo = repository.NewDBLinksRepository(db)
	case parsedFlags.GetLocalStorage() != "":
		linksRepo = repository.NewLocalMemoryLinksRepository(parsedFlags.GetLocalStorage())
	default:
		linksRepo = repository.NewInMemoryLinksRepository()
	}
	linksService := service.NewLinksService(linksRepo)
	linksHandler := handler.NewLinkHandler(linksService, parsedFlags)
	r := chi.NewRouter()

	r.Use(
		middleware.RequestID,
		middleware.RealIP,
		middleware.Logger,
		middleware.Recoverer,
		internalmiddleware.GZIPMiddleware,
	)

	r.Use(middleware.Timeout(parsedFlags.GetTimeout()))

	r.Route("/", func(r chi.Router) {
		r.Get("/{shortURL}", linksHandler.Get)
		r.Post("/", linksHandler.PostText)
		r.Route("/api", func(r chi.Router) {
			r.Post("/shorten", linksHandler.PostJSON)
		})
		r.Get("/ping", linksHandler.PingDB)
	})

	return r
}
