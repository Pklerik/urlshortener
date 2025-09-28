package model

import (
	"fmt"
	"strings"
)

// Requester interface provide request struct.
type Requester interface {
	String() string
}

// Request provide request for shortener.
type Request struct {
	URL string `json:"url"`
}

// String (req *Request) returns string representation of interface realization.
func (req *Request) String() string {
	return fmt.Sprintf("Request{URL: %s}", req.URL)
}

// ReqPostBatch provide batch contract.
type ReqPostBatch struct {
	CorrelationID string `json:"correlation_id"`
	LongURL       string `json:"original_url"`
}

// SlReqPostBatch provide slice of batch requests.
type SlReqPostBatch []ReqPostBatch

// String (reqSl *SlReqPostBatch) returns string representation of interface realization.
func (reqSl *SlReqPostBatch) String() string {
	var res = "["
	for _, req := range *reqSl {
		res += fmt.Sprintf("ReqPostBatch{CorrelationID: %s, LongURL: %s}", req.CorrelationID, req.LongURL)
	}

	return fmt.Sprint(reqSl, "]")
}

type ShortUrls []string

func (us *ShortUrls) String() string {
	return fmt.Sprintf(`Urls{"%s"}`, strings.Join(*us, `", "`))
}
