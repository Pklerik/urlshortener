// Package router provides functionality for setups and startup of server.
package router

import (
	"context"
	"fmt"
	"net/http"

	//nolint

	"github.com/Pklerik/urlshortener/internal/config"
	"github.com/Pklerik/urlshortener/internal/handler"
	"github.com/Pklerik/urlshortener/internal/logger"
	"github.com/Pklerik/urlshortener/internal/middleware"
	"github.com/Pklerik/urlshortener/internal/repository"
	dbrepo "github.com/Pklerik/urlshortener/internal/repository/db"
	"github.com/Pklerik/urlshortener/internal/repository/inmemory"
	"github.com/Pklerik/urlshortener/internal/repository/localfile"
	"github.com/Pklerik/urlshortener/internal/service/links"
	"github.com/go-chi/chi"
	chimiddleware "github.com/go-chi/chi/middleware"
)

// ConfigureRouter starts server with base configuration.
func ConfigureRouter(ctx context.Context, parsedFlags config.StartupFlagsParser) (http.Handler, error) {
	var linksRepo repository.LinksStorager

	r := chi.NewRouter()

	dbConf, err := parsedFlags.GetDatabaseConf()
	switch {
	case err == nil:
		logger.Sugar.Info("Used DB realization")

		linksRepo, err = dbrepo.NewDBLinksRepository(ctx, dbConf)
		if err != nil {
			logger.Sugar.Error(err)
			return r, fmt.Errorf("ConfigureRouter: %w", err)
		}
	case parsedFlags.GetLocalStorage() != "":
		logger.Sugar.Info("Used File realization")

		linksRepo = localfile.NewLocalMemoryLinksRepository(parsedFlags.GetLocalStorage())
	default:
		logger.Sugar.Info("Used InMemory realization")

		linksRepo = inmemory.NewInMemoryLinksRepository()
	}

	linksService := links.NewLinksService(linksRepo)
	linksHandler := handler.NewLinkHandler(linksService, parsedFlags)

	r.Use(
		chimiddleware.RequestID,
		chimiddleware.RealIP,
		chimiddleware.Logger,
		// middleware.Recoverer,
		middleware.GZIPMiddleware,
		middleware.AuthUser,
	)

	r.Use(chimiddleware.Timeout(parsedFlags.GetTimeout()))

	r.Route("/", func(r chi.Router) {
		r.Get("/{shortURL}", linksHandler.Get)
		r.Post("/", linksHandler.PostText)
		r.Route("/api", func(r chi.Router) {
			r.Route("/shorten", func(r chi.Router) {
				r.Post("/", linksHandler.PostJSON)
				r.Post("/batch", linksHandler.PostBatchJSON)
			})
			r.Route("/user", func(r chi.Router) {
				r.Get("/urls", linksHandler.GetUserLinks)
				r.Delete("/urls", linksHandler.DeleteUserLinks)
			})
		})
		r.Get("/ping", linksHandler.PingDB)
	})

	return r, nil
}
