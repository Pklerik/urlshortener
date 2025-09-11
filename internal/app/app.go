// Package app contain app function for startup.
package app

import (
	"context"
	"database/sql"
	"net/http"
	"os"
	"os/signal"
	"strconv"

	//nolint
	"syscall"

	"github.com/Pklerik/urlshortener/internal/config"
	"github.com/Pklerik/urlshortener/internal/logger"
	"github.com/Pklerik/urlshortener/internal/router"
	"github.com/Pklerik/urlshortener/migrations"
	"golang.org/x/sync/errgroup"
)

// StartApp - starts server app function.
func StartApp(parsedArgs config.StartupFlagsParser) {
	var (
		db  *sql.DB
		err error
	)
	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		c := make(chan os.Signal, 1) // we need to reserve to buffer size 1, so the notifier are not blocked
		signal.Notify(c, os.Interrupt, syscall.SIGTERM)

		<-c
		cancel()
	}()

	db, err = ConnectDB(parsedArgs)
	if err != nil {
		logger.Sugar.Errorf("Cant connect to db server: %w", err)
	}
	defer db.Close()

	logger.Sugar.Infof("SUCCESS connecting to db: %v", db.Stats())

	err = migrations.MakeMigrations(ctx, db, parsedArgs.GetDatabaseConf())
	if err != nil {
		logger.Sugar.Errorf("Cant connect to db server: %w", err)
		return
	}

	argPort := ":" + strconv.Itoa(parsedArgs.GetServerAddress().Port)
	logger.Sugar.Infof("Setup server with args: port: %s", argPort)
	httpServer := &http.Server{
		Addr:         argPort,
		Handler:      router.ConfigureRouter(parsedArgs, db),
		ReadTimeout:  parsedArgs.GetTimeout(),
		WriteTimeout: parsedArgs.GetTimeout(),
	}

	g, gCtx := errgroup.WithContext(ctx)
	g.Go(func() error {
		logger.Sugar.Infof("Starting server")
		return httpServer.ListenAndServe()
	})
	g.Go(func() error {
		<-gCtx.Done()
		logger.Sugar.Infof("Stopped serving new connections.")

		return httpServer.Shutdown(context.Background())
	})

	if err := g.Wait(); err != nil {
		logger.Sugar.Infof("exit reason: %s \n", err)
	}
}
