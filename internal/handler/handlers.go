// Package handler contains handling logic for all pages
package handler

import (
	"io"
	"net/http"

	"github.com/Pklerik/urlshortener/internal/handler/validators"
	"github.com/Pklerik/urlshortener/internal/repository"
	"github.com/Pklerik/urlshortener/internal/service"
)

// MainPage provide base shortener logic handler.
func MainPage(res http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodPost:
		validators.HeaderPlain(&res, req)

		body, err := io.ReadAll(req.Body)
		if err != nil {
			http.Error(res, `BadRequest`, http.StatusBadRequest)
		}
		defer req.Body.Close()

		short, err := service.ShortURL(body)
		if err != nil {
			http.Error(res, `BadRequest`, http.StatusBadRequest)
		}

		res.WriteHeader(http.StatusCreated)

		_, err = res.Write([]byte(req.Host + `/` + short))
		if err != nil {
			http.Error(res, `Unexpected exception: `, http.StatusInternalServerError)
		}

		return

	case http.MethodGet:
		validators.HeaderPlain(&res, req)

		long, ok := (*repository.MapShorts())[req.RequestURI[1:]]
		if !ok {
			http.Error(res, `BadRequest`, http.StatusBadRequest)
		}

		res.Header().Add("Location", long)
		res.WriteHeader(http.StatusTemporaryRedirect)

		return
	}

	http.Error(res, `BadRequest`, http.StatusBadRequest)
}
