// Package router provides functionality for setups and startup of server.
package router

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/Pklerik/urlshortener/internal/config"
	grpcHandler "github.com/Pklerik/urlshortener/internal/handler/grpc/handler"
	restHandler "github.com/Pklerik/urlshortener/internal/handler/rest"
	"github.com/Pklerik/urlshortener/internal/logger"
	"github.com/Pklerik/urlshortener/internal/middleware"
	"github.com/Pklerik/urlshortener/internal/repository"
	dbrepo "github.com/Pklerik/urlshortener/internal/repository/db"
	"github.com/Pklerik/urlshortener/internal/repository/inmemory"
	"github.com/Pklerik/urlshortener/internal/repository/localfile"
	"github.com/Pklerik/urlshortener/internal/service/links"
	"github.com/go-chi/chi"
	chimiddleware "github.com/go-chi/chi/middleware"

	"golang.org/x/net/http2/h2c"
)

// ConfigureRouter starts server with base configuration.
func ConfigureRouter(ctx context.Context, parsedFlags config.StartupFlagsParser) (http.Handler, error) {
	var (
		linksRepo repository.LinksRepository
		err       error
	)

	r := chi.NewRouter()

	linksRepo, err = chooseRepoRealization(ctx, parsedFlags)
	if err != nil {
		return r, fmt.Errorf("ConfigureRouter: %w", err)
	}

	linksService := links.NewLinksService(linksRepo, parsedFlags.GetSecretKey())

	authHandler := restHandler.NewAuthenticationHandler(linksService)
	linksHandler := restHandler.NewLinkHandler(linksService, authHandler, parsedFlags)
	auditHandler := restHandler.NewAuditor(parsedFlags, authHandler)

	gh := grpcHandler.NewUsersLinksHandler(linksService)

	// Add pprof routes
	r.Mount("/debug", chimiddleware.Profiler())

	r.Group(func(r chi.Router) {
		r.Use(
			chimiddleware.RequestID,
			chimiddleware.RealIP,
			chimiddleware.Logger,
			chimiddleware.Recoverer,
			middleware.GZIPMiddleware,
			authHandler.AuthUser,
			chimiddleware.Timeout(parsedFlags.GetTimeout()),
		)
		r.Route("/", func(r chi.Router) {
			r.Group(func(r chi.Router) {
				r.Use(auditHandler.AuditMiddleware)
				r.Post("/", linksHandler.PostText)
				r.Get("/{shortURL}", linksHandler.Get)
			})
			r.Route("/api", func(r chi.Router) {
				r.Route("/shorten", func(r chi.Router) {
					r.Group(func(r chi.Router) {
						r.Use(auditHandler.AuditMiddleware)
						r.Post("/", linksHandler.PostJSON)
					})
					r.Post("/batch", linksHandler.PostBatchJSON)
				})
				r.Route("/user", func(r chi.Router) {
					r.Get("/urls", linksHandler.GetUserLinks)
					r.Delete("/urls", linksHandler.DeleteUserLinks)
				})
				r.Route("/internal", func(r chi.Router) {
					r.Group(func(r chi.Router) {
						r.Use(middleware.TrustedSubnetMiddleware(parsedFlags))
						r.Get("/stats", linksHandler.GetStats)
					})
				})
			})
			r.Get("/ping", linksHandler.PingDB)
		})
	})

	r.Handle("/api/shorten"+"*", gh)   // e.g., mounts to "/protoPackage.ServiceName/"
	r.Handle("/{shortURL}"+"*", gh)    // e.g., mounts to "/protoPackage.ServiceName/"
	r.Handle("/api/user/urls"+"*", gh) // e.g., mounts to "/protoPackage.ServiceName/"

	printRoutes(r)

	return h2c.NewHandler(r, nil), nil
}

// Use chi.Walk to print all routes.
func printRoutes(r *chi.Mux) {
	err := chi.Walk(r, func(method string, route string, _ http.Handler, _ ...func(http.Handler) http.Handler) error {
		// Skip debug routes
		if strings.HasPrefix(route, "/debug") {
			return nil
		}

		logger.Sugar.Infof("[%s] %s", method, route)

		return nil
	})
	if err != nil {
		logger.Sugar.Error(err)
	}
}

func chooseRepoRealization(ctx context.Context, parsedFlags config.StartupFlagsParser) (repository.LinksRepository, error) {
	dbConf, err := parsedFlags.GetDatabaseConf()
	switch {
	case err == nil:
		logger.Sugar.Info("Used DB realization")

		repo, err := dbrepo.NewDBLinksRepository(ctx, dbConf)
		if err != nil {
			logger.Sugar.Error(err)
			return repo, fmt.Errorf("ConfigureRouter: %w", err)
		}

		return repo, nil
	case parsedFlags.GetLocalStorage() != "":
		logger.Sugar.Info("Used File realization")

		return localfile.NewLocalMemoryLinksRepository(parsedFlags.GetLocalStorage()), nil
	default:
		logger.Sugar.Info("Used InMemory realization")

		return inmemory.NewInMemoryLinksRepository(), nil
	}
}
