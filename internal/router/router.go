// Package router provides functionality for setups and startup of server.
package router

import (
	"errors"
	"log"
	"net/http"

	"github.com/Pklerik/urlshortener/internal/handler"
)

// StartServer starts server with base configuration.
func StartServer() {
	mux := http.NewServeMux()
	mux.HandleFunc(`/`, handler.MainPage)

	log.Println("Starting server")
	if err := http.ListenAndServe(`:8080`, mux); !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("HTTP server error: %v", err)
	}

	log.Println("Stopped serving new connections.")
}
