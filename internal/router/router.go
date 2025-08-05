// Package router provides functionality for setups and startup of server.
package router

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	//nolint
	"syscall"

	"github.com/Pklerik/urlshortener/internal/handler"
	"github.com/Pklerik/urlshortener/internal/repository"
	"github.com/Pklerik/urlshortener/internal/service"
	"golang.org/x/sync/errgroup"
)

// StartServer starts server with base configuration.
func StartServer() {
	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		c := make(chan os.Signal, 1) // we need to reserve to buffer size 1, so the notifier are not blocked
		signal.Notify(c, os.Interrupt, syscall.SIGTERM)

		<-c
		cancel()
	}()

	linksRepo := repository.NewInMemoryLinksRepository()
	linksService := service.NewLinksService(linksRepo)
	linksHandler := handler.NewLinkHandler(linksService)
	mux := http.NewServeMux()
	mux.HandleFunc(`/`, linksHandler.RegisterLinkHandler)

	httpServer := &http.Server{
		Addr:         ":8080",
		Handler:      mux,
		ReadTimeout:  10 * time.Minute,
		WriteTimeout: 10 * time.Minute,
	}

	g, gCtx := errgroup.WithContext(ctx)
	g.Go(func() error {
		log.Println("Starting server")
		return httpServer.ListenAndServe()
	})
	g.Go(func() error {
		<-gCtx.Done()
		log.Println("Stopped serving new connections.")

		return httpServer.Shutdown(context.Background())
	})

	if err := g.Wait(); err != nil {
		log.Printf("exit reason: %s \n", err)
	}
}
