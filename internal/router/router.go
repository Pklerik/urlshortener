package router

import (
	"errors"
	"log"
	"net/http"

	"github.com/Pklerik/urlshortener/internal/handler"
)

func StartServer() {
	mux := http.NewServeMux()
	mux.HandleFunc(`/`, handler.MainPage)
	mux.HandleFunc(`/id`, handler.IdPage)

	if err := http.ListenAndServe(`:8080`, mux); !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("HTTP server error: %v", err)
	}
	log.Println("Stopped serving new connections.")
}
