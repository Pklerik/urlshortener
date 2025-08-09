// Package app contain app function for startup.
package app

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"

	//nolint
	"syscall"
	"time"

	"github.com/Pklerik/urlshortener/internal/config"
	"github.com/Pklerik/urlshortener/internal/router"
	"golang.org/x/sync/errgroup"
)

// StartApp - starts server app function.
func StartApp(parsedArgs *config.StartupFlags) {
	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		c := make(chan os.Signal, 1) // we need to reserve to buffer size 1, so the notifier are not blocked
		signal.Notify(c, os.Interrupt, syscall.SIGTERM)

		<-c
		cancel()
	}()

	argPort := ":" + strconv.Itoa(parsedArgs.ServerAddress.Port)
	log.Printf("Setup server with args: port: %s", argPort)
	httpServer := &http.Server{
		Addr:         argPort,
		Handler:      router.ConfigureRouter(parsedArgs),
		ReadTimeout:  time.Duration(parsedArgs.Timeout.Seconds * float64(time.Second)),
		WriteTimeout: time.Duration(parsedArgs.Timeout.Seconds * float64(time.Second)),
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
