// Package app contain app function for startup.
package app

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Pklerik/urlshortener/internal/router"
	"golang.org/x/sync/errgroup"
)

// StartServer - starts server app function.
func StartServer() {
	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		c := make(chan os.Signal, 1) // we need to reserve to buffer size 1, so the notifier are not blocked
		signal.Notify(c, os.Interrupt, syscall.SIGTERM)

		<-c
		cancel()
	}()

	httpServer := &http.Server{
		Addr:         ":8080",
		Handler:      router.ConfigureRouter(),
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
