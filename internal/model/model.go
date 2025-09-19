// Package model contains structures for application and doesn't contain business logic.
package model

import (
	"errors"
	"fmt"
	"net/http"
)

var (
	ErrReadRequest = errors.New("unable to read request")
)

// LinkUUIDv7 is a custom type that embeds uuidv7.UUID.
type LinkUUIDv7 string

// LinkData provide structure for URLs storage.
type LinkData struct {
	UUID     LinkUUIDv7 `json:"uuid"`
	ShortURL string     `json:"short_url"`
	LongURL  string     `json:"original_url"`
}

func (ld *LinkData) String() string {
	return fmt.Sprintf(`LinkData{UUID: %s, ShortURL: %s, LongURL: %s}`, ld.UUID, ld.ShortURL, ld.LongURL)
}

type Requester interface {
	StringReq(r *http.Request) string
}

// Request provide request for shortener.
type Request struct {
	URL string `json:"url"`
}

func (req *Request) StringReq(r *http.Request) string {
	return fmt.Sprintf("Request{URL: %s}", req.URL)
}

// ReqPostBatch provide batch contract.
type ReqPostBatch struct {
	CorrelationID string `json:"correlation_id"`
	LongURL       string `json:"original_url"`
}

type SlReqPostBatch []ReqPostBatch

func (reqSl *SlReqPostBatch) StringReq(r *http.Request) string {
	var res string = "["
	for _, req := range *reqSl {
		res += fmt.Sprintf("ReqPostBatch{CorrelationID: %s, LongURL: %s}", req.CorrelationID, req.LongURL)
	}
	res += "]"
	return fmt.Sprint(reqSl)
}

type Responser interface {
	Response() string
}

// Response provide response for shortener.
type Response struct {
	Result string `json:"result"`
}

func (r *Response) Response() string {
	return fmt.Sprintf("Response{Result: %s}", r.Result)
}

// ResPostBatch provide batch contract.
type ResPostBatch struct {
	CorrelationID string `json:"correlation_id"`
	ShortURL      string `json:"short_url"`
}

func (r *ResPostBatch) Response() string {
	return fmt.Sprintf("Response{CorrelationID: %s, ShortURL: %s}", r.CorrelationID, r.ShortURL)
}

type SlResPostBatch []ResPostBatch

func (r *SlResPostBatch) Response() string {
	var res string = "["
	for _, resp := range *r {
		res += fmt.Sprintf("RespPostBatch{CorrelationID: %s, LongURL: %s}", resp.CorrelationID, resp.ShortURL)
	}
	res += "]"
	return res
}
