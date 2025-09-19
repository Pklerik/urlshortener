// Package router provides functionality for setups and startup of server.
package router

import (
	"context"
	"net/http"

	//nolint

	"github.com/Pklerik/urlshortener/internal/config"
	"github.com/Pklerik/urlshortener/internal/handler"
	"github.com/Pklerik/urlshortener/internal/internalmiddleware"
	"github.com/Pklerik/urlshortener/internal/logger"
	"github.com/Pklerik/urlshortener/internal/repository"
	"github.com/Pklerik/urlshortener/internal/service"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

// ConfigureRouter starts server with base configuration.
func ConfigureRouter(ctx context.Context, parsedFlags config.StartupFlagsParser) http.Handler {
	var linksRepo repository.LinksStorager

	dbConf, err := parsedFlags.GetDatabaseConf()
	switch {
	case err == nil:
		logger.Sugar.Info("Used DB realization")

		linksRepo = repository.NewDBLinksRepository(ctx, dbConf)
	case parsedFlags.GetLocalStorage() != "":
		logger.Sugar.Info("Used File realization")

		linksRepo = repository.NewLocalMemoryLinksRepository(parsedFlags.GetLocalStorage())
	default:
		logger.Sugar.Info("Used InMemory realization")

		linksRepo = repository.NewInMemoryLinksRepository()
	}

	linksService := service.NewLinksService(linksRepo)
	linksHandler := handler.NewLinkHandler(linksService, parsedFlags)
	r := chi.NewRouter()

	r.Use(
		middleware.RequestID,
		middleware.RealIP,
		middleware.Logger,
		// middleware.Recoverer,
		internalmiddleware.GZIPMiddleware,
	)

	r.Use(middleware.Timeout(parsedFlags.GetTimeout()))

	r.Route("/", func(r chi.Router) {
		r.Get("/{shortURL}", linksHandler.Get)
		r.Post("/", linksHandler.PostText)
		r.Route("/api", func(r chi.Router) {
			r.Route("/shorten", func(r chi.Router) {
				r.Post("/", linksHandler.PostJSON)
				r.Post("/batch", linksHandler.PostBatchJSON)
			})
		})
		r.Get("/ping", linksHandler.PingDB)
	})

	return r
}
