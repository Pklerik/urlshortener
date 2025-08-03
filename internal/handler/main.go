package handler

import (
	"bytes"
	"io"
	"net/http"
	"slices"
	"strings"

	"github.com/Pklerik/urlshortener/internal/repository"
	"github.com/Pklerik/urlshortener/internal/service"
)

func MainPage(res http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodPost:
		if value, ok := req.Header[`Content-Type`]; ok {
			if !slices.ContainsFunc(value, func(s string) bool { return strings.Contains(s, `text/plain`) }) {
				http.Error(res, `BadRequest`, http.StatusBadRequest)
			}
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
			if _, err := res.Write(short[:]); err != nil {
				http.Error(res, `Unexpected exception: `, http.StatusInternalServerError)
			}

		}
		return
	case http.MethodGet:
		if value, ok := req.Header[`Content-Type`]; ok {
			if !slices.Contains(value, "text/plain") {
				http.Error(res, `BadRequest`, http.StatusBadRequest)
			}
			shortID := [10]byte([]byte(req.RequestURI[1:]))
			if long, ok := repository.MapShortener[shortID]; ok {
				res.Header().Add("Location", string(long))
				res.WriteHeader(http.StatusTemporaryRedirect)
			}
		}
		return
	}
	http.Error(res, `BadRequest`, http.StatusBadRequest)
}

func IdPage(res http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		http.Error(res, `Method Not Allowed`, http.StatusMethodNotAllowed)
	}
	res.WriteHeader(http.StatusOK)
	buf := bytes.NewBuffer([]byte{})

	for key := range repository.MapShortener {
		buf.Write(key[:])
		buf.Write([]byte("\n"))
	}
	if _, err := res.Write(buf.Bytes()); err != nil {
		http.Error(res, `Unexpected exception`, http.StatusInternalServerError)
	}

}
