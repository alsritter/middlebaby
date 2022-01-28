package common

import (
	"math/rand"
	"time"
)

// Imposter define an imposter structure
type HttpImposter struct {
	Request  Request  `json:"request"`
	Response Response `json:"response"`
}

// Delay returns delay for response that user can specify in imposter config
func (i *HttpImposter) Delay() time.Duration {
	return i.Response.Delay.GetDelay()
}

// Request represent the structure of real request
type Request struct {
	Method  string            `json:"method"`
	Url     string            `json:"url"`
	Headers map[string]string `json:"headers"`
	Params  map[string]string `json:"params"`
}

// Response represent the structure of real response
type Response struct {
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

// Delay return random time.Duration with respect to specified time range.
func (d *ResponseDelay) GetDelay() time.Duration {
	offset := d.Offset
	if offset > 0 {
		offset = rand.Int63n(d.Offset)
	}
	return time.Duration(d.Delay+offset) * time.Millisecond
}
