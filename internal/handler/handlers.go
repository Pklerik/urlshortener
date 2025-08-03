// Package handler contains handling logic for all pages
package handler

import (
	"io"
	"log"
	"net/http"

	"github.com/Pklerik/urlshortener/internal/handler/validators"
	"github.com/Pklerik/urlshortener/internal/repository"
	"github.com/Pklerik/urlshortener/internal/service"
)

// MainPage provide base shortener logic handler.
func MainPage(res http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodPost:
		validators.TextPlain(&res, req)

		body, err := io.ReadAll(req.Body)
		if err != nil {
			log.Printf(`Unable to read body: status: %d`, http.StatusBadRequest)
			http.Error(res, `Unable to read body`, http.StatusBadRequest)
		}
		defer req.Body.Close()

		short, err := service.ShortURL(body)
		if err != nil {
			log.Printf(`Unable to shorten URL: status: %d`, http.StatusBadRequest)
			http.Error(res, `Unable to shorten URL`, http.StatusBadRequest)
		}

		res.WriteHeader(http.StatusCreated)

		_, err = res.Write([]byte(`http://` + req.Host + `/` + short))
		if err != nil {
			log.Printf(`Unexpected exception: status: %d`, http.StatusInternalServerError)
			http.Error(res, `Unexpected exception: `, http.StatusInternalServerError)
		}

		return

	case http.MethodGet:
		long, ok := (*repository.MapShorts())[req.RequestURI[1:]]
		if !ok {
			log.Printf(`Unable to find long URL for short: %s: status: %d`, req.RequestURI[1:], http.StatusBadRequest)
			http.Error(res, `Unable to find long URL for short`, http.StatusBadRequest)
		}

		log.Println("Map of short links: ", repository.MapShorts())
		res.Header().Add("Location", long)

		res.WriteHeader(http.StatusTemporaryRedirect)
		log.Println("Full header: ", res.Header())

		return
	}

	log.Printf(`Method not implemented: %s, status: %d`, req.Method, http.StatusNotImplemented)
	http.Error(res, `BadRequest`, http.StatusBadRequest)
}
