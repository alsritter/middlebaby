package proxy

import (
	"math/rand"
	"time"
)

// Request represent the structure of real request
type Request struct {
	Method  string             `json:"method"`
	Url     string             `json:"url"`
	Params  *map[string]string `json:"params"`
	Headers *map[string]string `json:"headers"`
}

// Response represent the structure of real response
type Response struct {
	Status  int                `json:"status"`
	Body    string             `json:"body"`
	Headers *map[string]string `json:"headers"`
	Delay   ResponseDelay      `json:"delay"`
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

// Imposter define an imposter structure
type Imposter struct {
	BasePath string
	Request  Request  `json:"request"`
	Response Response `json:"response"`
}

// Delay returns delay for response that user can specify in imposter config
func (i *Imposter) Delay() time.Duration {
	return i.Response.Delay.GetDelay()
}
