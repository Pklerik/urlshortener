package handler

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/Pklerik/urlshortener/internal/logger"
	"github.com/Pklerik/urlshortener/internal/model"
	"github.com/goccy/go-json"
	"go.uber.org/zap"
)

func readReq(r *http.Request, body []byte, req model.Requester) error {
	contentType := r.Header.Get("Content-Type")
	switch {
	case strings.Contains(contentType, "application/json"):
		reader := io.NopCloser(bytes.NewReader(body))

		dec := json.NewDecoder(reader)
		if err := dec.Decode(req); err != nil {
			logger.Log.Debug("cannot decode request JSON body", zap.Error(err))
			return fmt.Errorf("(req *Request) Read: cannot decode request JSON body: %w", err)
		}

		return nil
	case strings.Contains(contentType, "text/plain"):
		if reqText, ok := req.(model.RequestTextPlainHandler); ok {
			reqText.SetBody(string(body))
		}

		return nil
	}
	return nil

}

func writeRes(w http.ResponseWriter, res model.Responser) error {
	enc := json.NewEncoder(w)
	logger.Sugar.Debugf("Head: %v", w.Header())

	if err := enc.Encode(res); err != nil {
		return fmt.Errorf("writing Response type: %T, value: %v, error: %w ", res, res, err)
	}

	return nil
}
