package proxy

import (
	"math/rand"
	"time"
)

// Request represent the structure of real request
type Request struct {
	Method     string             `json:"method"`
	Endpoint   string             `json:"endpoint"`
	SchemaFile *string            `json:"schemaFile"`
	Params     *map[string]string `json:"params"`
	Headers    *map[string]string `json:"headers"`
}

// Response represent the structure of real response
type Response struct {
	Status   int                `json:"status"`
	Body     string             `json:"body"`
	BodyFile *string            `json:"bodyFile"`
	Headers  *map[string]string `json:"headers"`
	Delay    ResponseDelay      `json:"delay"`
}

// Imposter define an imposter structure
type Imposter struct {
	BasePath string
	Request  Request  `json:"request"`
	Response Response `json:"response"`
}

// ResponseDelay represent time delay before server responds.
type ResponseDelay struct {
	delay  int64
	offset int64
}

// Delay return random time.Duration with respect to specified time range.
func (d *ResponseDelay) Delay() time.Duration {
	offset := d.offset
	if offset > 0 {
		offset = rand.Int63n(d.offset)
	}
	return time.Duration(d.delay + offset)
}
