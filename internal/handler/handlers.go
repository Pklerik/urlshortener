// Package handler contains handling logic for all pages
package handler

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/goccy/go-json"

	"github.com/Pklerik/urlshortener/internal/config"
	"github.com/Pklerik/urlshortener/internal/handler/validators"
	"github.com/Pklerik/urlshortener/internal/logger"

	middleware "github.com/Pklerik/urlshortener/internal/middleware"
	"github.com/Pklerik/urlshortener/internal/model"
	"github.com/Pklerik/urlshortener/internal/repository"
	"github.com/Pklerik/urlshortener/internal/service"
	"github.com/go-chi/chi"
	"go.uber.org/zap"
)

// LinkHandler - provide contract for request handling.
type LinkHandler interface {
	Get(w http.ResponseWriter, r *http.Request)
	PostText(w http.ResponseWriter, r *http.Request)
	PostJSON(w http.ResponseWriter, r *http.Request)
	PingDB(w http.ResponseWriter, r *http.Request)
	PostBatchJSON(w http.ResponseWriter, r *http.Request)
	GetUserLinks(w http.ResponseWriter, r *http.Request)
	DeleteUserLinks(w http.ResponseWriter, r *http.Request)
}

// LinkHandle - wrapper for service handling.
type LinkHandle struct {
	service service.LinkServicer
	Args    config.StartupFlagsParser
}

// NewLinkHandler returns instance of LinkHandler.
func NewLinkHandler(userService service.LinkServicer, args config.StartupFlagsParser) LinkHandler {
	return &LinkHandle{service: userService, Args: args}
}

// Get returns Handler for URLs registration for GET method.
func (lh *LinkHandle) Get(w http.ResponseWriter, r *http.Request) {
	logger.Sugar.Infof(`Full request: %#v`, *r)

	ld, err := lh.service.GetShort(r.Context(), chi.URLParam(r, "shortURL"))
	if err != nil {
		logger.Sugar.Infof(`Unable to find long URL for short: %s: status: %d`, r.URL.Path[1:], http.StatusBadRequest)
		http.Error(w, `Unable to find long URL for short`, http.StatusBadRequest)

		return
	}

	w.Header().Add("Location", ld.LongURL)

	w.WriteHeader(http.StatusTemporaryRedirect)
	logger.Sugar.Infof(`Full Link: %s, for Short "%s"`, ld.LongURL, chi.URLParam(r, "shortURL"))
}

// PostText returns Handler for URLs registration for GET method.
func (lh *LinkHandle) PostText(w http.ResponseWriter, r *http.Request) {
	err := validators.TextPlain(w, r)
	if err != nil {
		logger.Sugar.Debugf("Unsupported Content-type: Header: after validation:", w.Header().Values("Content-type"))
		return
	}

	defer r.Body.Close()

	body, err := io.ReadAll(r.Body)
	if err != nil {
		logger.Sugar.Infof(`Unable to read body: status: %d`, http.StatusBadRequest)
		http.Error(w, `Unable to read body`, http.StatusBadRequest)

		return
	}

	userID, err := middleware.GetUserIDFromCookie(w, r)
	if err != nil {
		return
	}

	lds, err := lh.service.RegisterLinks(r.Context(), []string{string(body)}, userID)
	if err != nil && !errors.Is(err, repository.ErrExistingLink) {
		logger.Sugar.Infof(`Unable to shorten URL: status: %d`, http.StatusBadRequest)
		http.Error(w, `Unable to shorten URL`, http.StatusBadRequest)

		return
	}

	if errors.Is(err, repository.ErrExistingLink) {
		logger.Sugar.Infof(`Found existing urls: status: %d`, http.StatusConflict)
		w.WriteHeader(http.StatusConflict)
	} else {
		w.WriteHeader(http.StatusCreated)
	}

	redirectURL := lh.Args.GetAddressShortURL() + "/" + lds[0].ShortURL

	_, err = w.Write([]byte(redirectURL))
	if err != nil {
		logger.Sugar.Infof(`Unexpected exception: status: %w`, err)
		http.Error(w, `Unexpected exception: `, http.StatusInternalServerError)

		return
	}

	logger.Sugar.Infof(`created ShortURL redirection: "%s" for longURL: "%s"`, redirectURL, lds[0].LongURL)
}

// PostJSON returns Handler for URLs registration for GET method.
func (lh *LinkHandle) PostJSON(w http.ResponseWriter, r *http.Request) {
	err := validators.ApplicationJSON(w, r)
	if err != nil {
		return
	}

	var req model.Request
	if err := readReq(r, &req); err != nil {
		logger.Log.Debug("cannot read request", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
	}

	userID, err := middleware.GetUserIDFromCookie(w, r)
	if err != nil {
		return
	}

	lds, err := lh.service.RegisterLinks(r.Context(), []string{req.URL}, userID)
	if err != nil && !errors.Is(err, repository.ErrExistingLink) {
		logger.Sugar.Infof(`Unable to shorten URL: status: %d`, http.StatusBadRequest)
		http.Error(w, `Unable to shorten URL`, http.StatusBadRequest)

		return
	}

	if errors.Is(err, repository.ErrExistingLink) {
		logger.Sugar.Infof(`Found existing urls: status: %d`, http.StatusConflict)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusConflict)
	} else {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
	}

	if len(lds) > 1 {
		http.Error(w, "Not implemented multiple response", http.StatusInternalServerError)

		return
	}

	resp := model.Response{
		Result: lh.Args.GetAddressShortURL() + "/" + lds[0].ShortURL,
	}

	if err := writeRes(w, &resp); err != nil {
		logger.Log.Debug("error encoding response", zap.Error(err))
		http.Error(w, `Unexpected exception: `, http.StatusInternalServerError)

		return
	}

	logger.Sugar.Infof(`created ShortURL redirection: "%s" for longURL: "%s"`, resp.Result, lds[0].LongURL)
}

// PingDB provide 200 for successful database ping.
func (lh *LinkHandle) PingDB(w http.ResponseWriter, r *http.Request) {
	logger.Sugar.Infof(`Full request: %#v`, *r)

	ctx, cancel := context.WithTimeout(context.Background(), lh.Args.GetTimeout())
	defer cancel()

	if err := lh.service.PingDB(ctx); err != nil {
		http.Error(w, "ping db error", http.StatusInternalServerError)

		return
	}

	w.WriteHeader(http.StatusOK)
}

// PostBatchJSON provide json batch POST new links realization.
func (lh *LinkHandle) PostBatchJSON(w http.ResponseWriter, r *http.Request) {
	err := validators.ApplicationJSON(w, r)
	if err != nil {
		return
	}

	var req model.SlReqPostBatch
	if err := readReq(r, &req); err != nil {
		logger.Log.Debug("cannot read request", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
	}

	logger.Sugar.Infof("req struct for batch: %s", req)

	reqLongUrls := make([]string, 0, len(req))
	for _, reqElem := range req {
		reqLongUrls = append(reqLongUrls, reqElem.LongURL)
	}

	userID, err := middleware.GetUserIDFromCookie(w, r)
	if err != nil {
		return
	}

	lds, err := lh.service.RegisterLinks(r.Context(), reqLongUrls, userID)
	if err != nil && !errors.Is(err, repository.ErrExistingLink) {
		logger.Sugar.Infof(`Unable to shorten URL: status: %d`, http.StatusBadRequest)
		http.Error(w, `Unable to shorten URL`, http.StatusBadRequest)

		return
	}

	resp := make(model.SlResPostBatch, 0, len(lds))
	for i, linkData := range lds {
		resp = append(resp, model.ResPostBatch{
			CorrelationID: req[i].CorrelationID,
			ShortURL:      lh.Args.GetAddressShortURL() + "/" + linkData.ShortURL,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	if err := writeRes(w, &resp); err != nil {
		logger.Log.Debug("error encoding response", zap.Error(err))
		http.Error(w, `Unexpected exception: `, http.StatusInternalServerError)

		return
	}
}

// GetUserLinks for handle get request for user data.
func (lh *LinkHandle) GetUserLinks(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.GetUserIDFromCookie(w, r)
	if err != nil {
		return
	}

	lds, err := lh.service.ProvideUserLinks(r.Context(), userID)
	if err != nil && !errors.Is(err, repository.ErrNotFoundLink) {
		logger.Sugar.Infof(`Unable to get URLs for User %d: status: %d`, userID, http.StatusBadRequest)
		http.Error(w, fmt.Sprintf(`Unable to get URLs for User %d: status: %d`, userID, http.StatusBadRequest), http.StatusBadRequest)

		return
	}

	if errors.Is(err, repository.ErrNotFoundLink) {
		logger.Sugar.Infof(`No links found for User %d: status: %d`, userID, http.StatusBadRequest)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNoContent)

		return
	}

	resp := make(model.LongShortURLs, 0, len(lds))
	for _, linkData := range lds {
		resp = append(resp, model.LongShortURL{
			LongURL:  linkData.LongURL,
			ShortURL: lh.Args.GetAddressShortURL() + "/" + linkData.ShortURL,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := writeRes(w, &resp); err != nil {
		logger.Log.Debug("error encoding response", zap.Error(err))
		http.Error(w, `Unexpected exception: `, http.StatusInternalServerError)

		return
	}
}

func (lh *LinkHandle) DeleteUserLinks(w http.ResponseWriter, r *http.Request) {
	err := validators.ApplicationJSON(w, r)
	if err != nil {
		return
	}

	var req model.ShortUrls
	if err := readReq(r, &req); err != nil {
		logger.Log.Debug("cannot read request", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
	}

	userID, err := middleware.GetUserIDFromCookie(w, r)
	if err != nil {
		return
	}

	// Асинхронное выполнение
	go func() {
		if _, err := lh.service.MarkAsDeleted(r.Context(), userID, req); err != nil {
			logger.Sugar.Errorf("Failed to mark URLs as deleted: %v", err)
		}
	}()

	if err != nil {

	}
	w.WriteHeader(http.StatusAccepted)
	logger.Sugar.Infof(`url: "%s" Accepted for deletion`, req)

}

func readReq(r *http.Request, req model.Requester) error {
	defer r.Body.Close()

	body, err := io.ReadAll(r.Body)
	if err != nil {
		logger.Log.Debug("cannot read body", zap.Error(err))
		return fmt.Errorf("(req *Request) Read: cannot read body: %w", err)
	}

	reader := io.NopCloser(bytes.NewReader(body))

	dec := json.NewDecoder(reader)
	if err := dec.Decode(req); err != nil {
		logger.Log.Debug("cannot decode request JSON body", zap.Error(err))
		return fmt.Errorf("(req *Request) Read: cannot decode request JSON body: %w", err)
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
