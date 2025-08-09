// Package handler contains handling logic for all pages
package handler

import (
	"io"
	"log"
	"net/http"

	"github.com/Pklerik/urlshortener/internal/config"
	"github.com/Pklerik/urlshortener/internal/handler/validators"
	"github.com/Pklerik/urlshortener/internal/service"
	"github.com/go-chi/chi"
)

// LinkHandler - provide contract for request handling.
type LinkHandler interface {
	Get(w http.ResponseWriter, r *http.Request)
	Post(w http.ResponseWriter, r *http.Request)
}

// LinkHandle - wrapper for service handling.
type LinkHandle struct {
	linkService service.LinkServicer
	Args        *config.StartupFlags
}

// NewLinkHandler returns instance of LinkHandler.
func NewLinkHandler(userService service.LinkServicer, args *config.StartupFlags) LinkHandler {
	return &LinkHandle{linkService: userService, Args: args}
}

// Get returns Handler for URLs registration for GET method.
func (lh *LinkHandle) Get(w http.ResponseWriter, r *http.Request) {
	log.Printf(`Full request: %#v`, *r)

	ld, err := lh.linkService.GetShort(r.Context(), chi.URLParam(r, "shortURL"))
	if err != nil {
		log.Printf(`Unable to find long URL for short: %s: status: %d`, r.URL.Path[1:], http.StatusBadRequest)
		http.Error(w, `Unable to find long URL for short`, http.StatusBadRequest)

		return
	}

	w.Header().Add("Location", ld.LongURL)

	w.WriteHeader(http.StatusTemporaryRedirect)
	log.Printf(`Full Link: %s, for Short "%s"`, ld.LongURL, chi.URLParam(r, "shortURL"))
}

// Post returns Handler for URLs registration for GET method.
func (lh *LinkHandle) Post(w http.ResponseWriter, r *http.Request) {
	err := validators.TextPlain(w, r)
	if err != nil {
		return
	}

	defer r.Body.Close()

	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf(`Unable to read body: status: %d`, http.StatusBadRequest)
		http.Error(w, `Unable to read body`, http.StatusBadRequest)

		return
	}

	ld, err := lh.linkService.RegisterLink(r.Context(), string(body))
	if err != nil {
		log.Printf(`Unable to shorten URL: status: %d`, http.StatusBadRequest)
		http.Error(w, `Unable to shorten URL`, http.StatusBadRequest)

		return
	}

	w.WriteHeader(http.StatusCreated)

	redirectURL := lh.Args.AddressShortURL + "/" + ld.ShortURL

	_, err = w.Write([]byte(redirectURL))
	if err != nil {
		log.Printf(`Unexpected exception: status: %d`, http.StatusInternalServerError)
		http.Error(w, `Unexpected exception: `, http.StatusInternalServerError)

		return
	}

	log.Printf(`created ShortURL redirection: "%s" for longURL: "%s"`, redirectURL, ld.LongURL)
}
