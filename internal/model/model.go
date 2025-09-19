// Package model contains structures for application and doesn't contain business logic.
package model

import (
	"fmt"
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

// Requester interface provide request struct.
type Requester interface {
	StringReq() string
}

// Request provide request for shortener.
type Request struct {
	URL string `json:"url"`
}

// StringReq (req *Request) returns string representation of interface realization.
func (req *Request) StringReq() string {
	return fmt.Sprintf("Request{URL: %s}", req.URL)
}

// ReqPostBatch provide batch contract.
type ReqPostBatch struct {
	CorrelationID string `json:"correlation_id"`
	LongURL       string `json:"original_url"`
}

// SlReqPostBatch provide slice of batch requests.
type SlReqPostBatch []ReqPostBatch

// StringReq (reqSl *SlReqPostBatch) returns string representation of interface realization.
func (reqSl *SlReqPostBatch) StringReq() string {
	var res = "["
	for _, req := range *reqSl {
		res += fmt.Sprintf("ReqPostBatch{CorrelationID: %s, LongURL: %s}", req.CorrelationID, req.LongURL)
	}

	return fmt.Sprint(reqSl, "]")
}

// Responser interface provide response struct.
type Responser interface {
	Response() string
}

// Response provide response for shortener.
type Response struct {
	Result string `json:"result"`
}

// Response (r *Response) returns string representation of interface realization.
func (r *Response) Response() string {
	return fmt.Sprintf("Response{Result: %s}", r.Result)
}

// ResPostBatch provide batch contract.
type ResPostBatch struct {
	CorrelationID string `json:"correlation_id"`
	ShortURL      string `json:"short_url"`
}

// Response (r *SlResPostBatch) returns string representation of interface realization.
func (r *ResPostBatch) Response() string {
	return fmt.Sprintf("Response{CorrelationID: %s, ShortURL: %s}", r.CorrelationID, r.ShortURL)
}

// SlResPostBatch provide slice for batch contract.
type SlResPostBatch []ResPostBatch

// Response (r *SlResPostBatch) returns string representation of interface realization.
func (r *SlResPostBatch) Response() string {
	var res = "["
	for _, resp := range *r {
		res += fmt.Sprintf("RespPostBatch{CorrelationID: %s, LongURL: %s}", resp.CorrelationID, resp.ShortURL)
	}

	return fmt.Sprint(res, "]")
}
