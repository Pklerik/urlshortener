// Package handler contains handling logic for all pages
package handler

import (
	"io"
	"log"
	"net/http"

	"github.com/Pklerik/urlshortener/internal/handler/validators"
	"github.com/Pklerik/urlshortener/internal/service"
)

// LinkHandler - wrapper for service handling.
type LinkHandler struct {
	linkService *service.LinkService
}

// NewLinkHandler returns instance of LinkHandler.
func NewLinkHandler(userService *service.LinkService) *LinkHandler {
	return &LinkHandler{linkService: userService}
}

// RegisterLinkHandler returns Handler for URLs registration.
func (lh *LinkHandler) RegisterLinkHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		validators.TextPlain(w, r)

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

		_, err = w.Write([]byte(`http://` + r.Host + `/` + ld.ShortURL))
		if err != nil {
			log.Printf(`Unexpected exception: status: %d`, http.StatusInternalServerError)
			http.Error(w, `Unexpected exception: `, http.StatusInternalServerError)
		}

		return

	case http.MethodGet:
		ld, err := lh.linkService.GetShort(r.Context(), r.URL.Path[1:])
		if err != nil {
			log.Printf(`Unable to find long URL for short: %s: status: %d`, r.URL.Path[1:], http.StatusBadRequest)
			http.Error(w, `Unable to find long URL for short`, http.StatusBadRequest)
		}

		w.Header().Add("Location", ld.LongURL)

		w.WriteHeader(http.StatusTemporaryRedirect)
		log.Println("Full header: ", w.Header())

		return
	}

	log.Printf(`Method not implemented: %s, status: %d`, r.Method, http.StatusNotImplemented)
	http.Error(w, `BadRequest`, http.StatusBadRequest)
}
