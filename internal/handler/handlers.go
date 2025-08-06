// Package handler contains handling logic for all pages
package handler

import (
	"io"
	"log"
	"net/http"
	"strconv"

	"github.com/Pklerik/urlshortener/internal/config"
	"github.com/Pklerik/urlshortener/internal/handler/validators"
	"github.com/Pklerik/urlshortener/internal/service"
	"github.com/go-chi/chi"
)

// LinkHandler - wrapper for service handling.
type LinkHandler struct {
	linkService *service.LinkService
	Args        *config.StartupFalgs
}

// NewLinkHandler returns instance of LinkHandler.
func NewLinkHandler(userService *service.LinkService, args *config.StartupFalgs) *LinkHandler {
	return &LinkHandler{linkService: userService, Args: args}
}

// GetRegisterLinkHandler returns Handler for URLs registration for GET method.
func (lh *LinkHandler) GetRegisterLinkHandler(w http.ResponseWriter, r *http.Request) {
	ld, err := lh.linkService.GetShort(r.Context(), chi.URLParam(r, "shortURL"))
	if err != nil {
		log.Printf(`Unable to find long URL for short: %s: status: %d`, r.URL.Path[1:], http.StatusBadRequest)
		http.Error(w, `Unable to find long URL for short`, http.StatusBadRequest)
	}

	w.Header().Add("Location", ld.LongURL)

	w.WriteHeader(http.StatusTemporaryRedirect)
	log.Printf("Full header: %v", w.Header())
}

// PostRegisterLinkHandler returns Handler for URLs registration for GET method.
func (lh *LinkHandler) PostRegisterLinkHandler(w http.ResponseWriter, r *http.Request) {
	err := validators.TextPlain(w, r)
	if err != nil {
		return
	}

	defer r.Body.Close()

	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf(`Unable to read body: status: %d`, http.StatusBadRequest)
		http.Error(w, `Unable to read body`, http.StatusBadRequest)
	}

	ld, err := lh.linkService.RegisterLink(r.Context(), string(body))
	if err != nil {
		log.Printf(`Unable to shorten URL: status: %d`, http.StatusBadRequest)
		http.Error(w, `Unable to shorten URL`, http.StatusBadRequest)
	}

	w.WriteHeader(http.StatusCreated)

	_, err = w.Write([]byte(lh.Args.AddressShortURL.Protocol + "://" +
		lh.Args.AddressShortURL.Host + ":" +
		strconv.Itoa(lh.Args.AddressShortURL.Port) +
		`/` + ld.ShortURL))
	if err != nil {
		log.Printf(`Unexpected exception: status: %d`, http.StatusInternalServerError)
		http.Error(w, `Unexpected exception: `, http.StatusInternalServerError)
	}
}
