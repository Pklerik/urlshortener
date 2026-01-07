// Package router provides functionality for setups and startup of server.
package router

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strings"

	"github.com/Pklerik/urlshortener/internal/config"
	grpcHandler "github.com/Pklerik/urlshortener/internal/handler/grpc/handler"
	restHandler "github.com/Pklerik/urlshortener/internal/handler/rest"
	"github.com/Pklerik/urlshortener/internal/interceptor"
	"github.com/Pklerik/urlshortener/internal/logger"
	"github.com/Pklerik/urlshortener/internal/middleware"
	"github.com/Pklerik/urlshortener/internal/repository"
	dbrepo "github.com/Pklerik/urlshortener/internal/repository/db"
	"github.com/Pklerik/urlshortener/internal/repository/inmemory"
	"github.com/Pklerik/urlshortener/internal/repository/localfile"
	"github.com/Pklerik/urlshortener/internal/service/links"
	"github.com/go-chi/chi"
	chimiddleware "github.com/go-chi/chi/middleware"
	"google.golang.org/grpc"

	"golang.org/x/net/http2/h2c"
)

var (
	ErrEmptyTrustedIPs       = errors.New("empty trusted IP list")
	ErrUnableParseTrustedIPs = errors.New("unable to parse trusted IP list")
)

type Handlers struct {
	AuthHandler  restHandler.IAuthentication
	LinksHandler restHandler.LinkHandler
	AuditHandler restHandler.Auditer
	GRPCHandler  *grpc.Server
}

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

	hs := Handlers{
		AuthHandler:  authHandler,
		LinksHandler: restHandler.NewLinkHandler(linksService, authHandler, parsedFlags),
		AuditHandler: restHandler.NewAuditor(parsedFlags, authHandler),
		GRPCHandler:  grpcHandler.NewUsersLinksHandler(ctx, linksService).Register(interceptor.AuthUnaryServerInterceptor(parsedFlags.GetSecretKey())),
	}
	trustedIPNet, err := parseTrustedCIDR(parsedFlags)
	if err != nil {
		return r, fmt.Errorf("ConfigureRouter: %w", err)
	}

	r = addRESTRoutes(r, parsedFlags, hs, trustedIPNet)

	printRoutes(r)

	// return http2 supported handler for gRPC realization inside chi routing.
	return h2c.NewHandler(r, nil), nil
}

func addRESTRoutes(r *chi.Mux, parsedFlags config.StartupFlagsParser,
	hs Handlers, trustedIPNet *net.IPNet) *chi.Mux {
	r.Use(middleware.GRPCMuxMiddleware(hs.GRPCHandler))
	// Add pprof routes
	r.Mount("/debug", chimiddleware.Profiler())

	r.Group(func(r chi.Router) {
		r.Use(
			chimiddleware.RequestID,
			chimiddleware.RealIP,
			chimiddleware.Logger,
			chimiddleware.Recoverer,
			middleware.GZIPMiddleware,
			hs.AuthHandler.AuthUser,
			chimiddleware.Timeout(parsedFlags.GetTimeout()),
		)
		r.Route("/", func(r chi.Router) {
			r.Group(func(r chi.Router) {
				r.Use(hs.AuditHandler.AuditMiddleware)
				r.Post("/", hs.LinksHandler.PostText)
				r.Get("/{shortURL}", hs.LinksHandler.Get)
			})
			r.Route("/api", func(r chi.Router) {
				r.Route("/shorten", func(r chi.Router) {
					r.Group(func(r chi.Router) {
						r.Use(hs.AuditHandler.AuditMiddleware)
						r.Post("/", hs.LinksHandler.PostJSON)
					})
					r.Post("/batch", hs.LinksHandler.PostBatchJSON)
				})
				r.Route("/user", func(r chi.Router) {
					r.Get("/urls", hs.LinksHandler.GetUserLinks)
					r.Delete("/urls", hs.LinksHandler.DeleteUserLinks)
				})
				r.Route("/internal", func(r chi.Router) {
					r.Group(func(r chi.Router) {
						r.Use(middleware.TrustedSubnetMiddleware(trustedIPNet))
						r.Get("/stats", hs.LinksHandler.GetStats)
					})
				})
			})
			r.Get("/ping", hs.LinksHandler.PingDB)
		})
	})

	return r
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

func parseTrustedCIDR(parsedFlags config.StartupFlagsParser) (*net.IPNet, error) {
	v := parsedFlags.GetTrustedCIDR()
	if v == "" {
		logger.Sugar.Warn("no authorized IPs was provided")
		return &net.IPNet{}, ErrEmptyTrustedIPs
	}

	_, trustedIPNet, err := net.ParseCIDR(v)
	if err != nil {
		logger.Sugar.Errorf("unable to parse trusted CIDR: %w", err)

		return &net.IPNet{}, ErrUnableParseTrustedIPs
	}
	return trustedIPNet, nil

}
