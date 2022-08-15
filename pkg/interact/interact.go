package interact

import (
	"math/rand"
	"time"
)

// HttpImposter define an imposter structure
type HttpImposter struct {
	Id       string       `json:"-"`
	Request  HttpRequest  `json:"request"`
	Response HttpResponse `json:"response"`
}

// Delay returns delay for response that user can specify in imposter config
func (i *HttpImposter) Delay() time.Duration {
	return i.Response.Delay.GetDelay()
}

// HttpRequest represent the structure of real request
type HttpRequest struct {
	Method  string            `json:"method"`
	Url     string            `json:"url"`
	Headers map[string]string `json:"headers"`
	Params  map[string]string `json:"params"`
}

// HttpResponse represent the structure of real response
type HttpResponse struct {
	Status  int               `json:"status"`
	Headers map[string]string `json:"headers"`
	Body    string            `json:"body"`
	Delay   ResponseDelay     `json:"delay"`
}

// ResponseDelay represent time delay before server responds.
type ResponseDelay struct {
	Delay  int64 `json:"delay"`
	Offset int64 `json:"offset"`
}

// GetDelay return random time.Duration with respect to specified time range.
func (d *ResponseDelay) GetDelay() time.Duration {
	offset := d.Offset
	if offset > 0 {
		offset = rand.Int63n(d.Offset)
	}
	return time.Duration(d.Delay+offset) * time.Millisecond
}

// GRpcImposter TODO: fill in the details.
type GRpcImposter struct {
	Id       string       `json:"-"`
	Request  GRpcRequest  `json:"request"`
	Response GRpcResponse `json:"response"`
}

type GRpcRequest struct {
	Method  string            `json:"method"`
	Headers map[string]string `json:"headers"`
}

type GRpcResponse struct {
	Headers map[string]string `json:"headers"`
}
