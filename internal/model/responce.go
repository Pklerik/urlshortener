package model

import "fmt"

// Responser interface provide response struct.
type Responser interface {
	String() string
}

// Response provide response for shortener.
type Response struct {
	Result string `json:"result"`
}

// String (r *Response) returns string representation of interface realization.
func (r *Response) String() string {
	return fmt.Sprintf("Response{Result: %s}", r.Result)
}

// ResPostBatch provide batch contract.
type ResPostBatch struct {
	CorrelationID string `json:"correlation_id"`
	ShortURL      string `json:"short_url"`
}

// String (r *SlResPostBatch) returns string representation of interface realization.
func (r *ResPostBatch) String() string {
	return fmt.Sprintf("ResPostBatch{CorrelationID: %s, ShortURL: %s}", r.CorrelationID, r.ShortURL)
}

// SlResPostBatch provide slice for batch contract.
type SlResPostBatch []ResPostBatch

// String (r *SlResPostBatch) returns string representation of interface realization.
func (r *SlResPostBatch) String() string {
	var res = ""
	for _, resp := range *r {
		res += fmt.Sprintf("RespPostBatch{CorrelationID: %s, LongURL: %s}", resp.CorrelationID, resp.ShortURL)
	}

	return fmt.Sprint("[", res, "]")
}

// LongShortURL provide user links contract.
type LongShortURL struct {
	ShortURL string `json:"short_url"`
	LongURL  string `json:"original_url"`
}

// LongShortURLs provide slice for user links contract.
type LongShortURLs []LongShortURL

// String (lsu *LongShortURLs) returns string representation of interface realization.
func (lsu *LongShortURLs) String() string {
	var res string
	for _, resp := range *lsu {
		res += fmt.Sprintf("LongShortURL{ShortURL: %s, LongURL: %s}", resp.ShortURL, resp.ShortURL)
	}

	return fmt.Sprint("[", res, "]")
}
