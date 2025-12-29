// Package app contain app function for startup.
package app

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"

	//nolint необходимо получать SIGTERM для остановки процесса.
	"syscall"

	"github.com/Pklerik/urlshortener/internal/config"
	"github.com/Pklerik/urlshortener/internal/criptography"
	"github.com/Pklerik/urlshortener/internal/logger"
	"github.com/Pklerik/urlshortener/internal/router"
	"golang.org/x/sync/errgroup"
)

// StartApp - starts server app function.
func StartApp(parsedArgs config.StartupFlagsParser) {
	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		c := make(chan os.Signal, 1) // we need to reserve to buffer size 1, so the notifier are not blocked
		signal.Notify(c, os.Interrupt, syscall.SIGTERM)

		<-c
		cancel()
	}()

	argPort := ":" + strconv.Itoa(parsedArgs.GetServerAddress().Port)
	logger.Sugar.Infof("Setup server with args: port: %s", argPort)

	routerHandler, err := router.ConfigureRouter(ctx, parsedArgs)
	if err != nil {
		logger.Sugar.Errorf("Unable to start server: %w", err)
		return
	}

	httpServer := &http.Server{
		Addr:         argPort,
		Handler:      routerHandler,
		ReadTimeout:  parsedArgs.GetTimeout(),
		WriteTimeout: parsedArgs.GetTimeout(),
	}

	g, gCtx := errgroup.WithContext(ctx)
	g.Go(func() error {
		if parsedArgs.GetTLS() {
			return runTLSListener(httpServer)
		}

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

func runTLSListener(httpServer *http.Server) error {
	// Сохраняем сертификат и приватный ключ в файлы ../../../cert/cert.pem и ../../../cert/private.pem
	certPath, err := os.Executable()
	if err != nil {
		logger.Sugar.Errorf("unable o start server %v", err)
		return fmt.Errorf("unable o start server %w", err)
	}

	certPath = filepath.Join(filepath.Dir(filepath.Dir(filepath.Dir(certPath))), "cert")

	keys, err := criptography.GetSertKey(certPath)
	if err != nil {
		logger.Sugar.Errorf("unable to generate cert sequence err: %v", err)
	}

	logger.Sugar.Infof("Starting server with TLS")

	return fmt.Errorf("unable to start server with TLS: %w", httpServer.ListenAndServeTLS(keys.CertPEMFile, keys.PrivateKeyPEMFile))
}
