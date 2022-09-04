package interact

import (
	"encoding/json"
	"math/rand"
	"reflect"
	"time"
)

// Protocol defines the protocol of request
type Protocol string

// defines a set of known protocols
const (
	ProtocolHTTP Protocol = "HTTP"
	ProtocolGRPC Protocol = "GRPC"
)

// ImposterCase define an imposter structure (a mock case)
type ImposterCase struct {
	Id       string   `json:"-"`
	Request  Request  `json:"request" yaml:"request"`
	Response Response `json:"response" yaml:"response"`
}

// Delay returns delay for response that user can specify in imposter config
func (i *ImposterCase) Delay() time.Duration {
	return i.Response.Delay.GetDelay()
}

// Request defines the request structure
type Request struct {
	Protocol Protocol            `json:"protocol" yaml:"protocol"`
	Method   string              `json:"method" yaml:"method"`
	Host     string              `json:"host" yaml:"host"`
	Path     string              `json:"path" yaml:"path"`
	Header   map[string][]string `json:"header" yaml:"header"`
	Params   map[string]string   `json:"params" yaml:"params"`
	Body     interface{}         `json:"body" yaml:"body"`
}

func (r *Request) GetBodyString() string {
	if r.Body != nil {
		if reflect.TypeOf(r.Body).Kind() == reflect.String {
			return r.Body.(string)
		}

		if _, ok := r.Body.([]byte); ok {
			return string(r.Body.([]byte))
		}

		str, _ := json.Marshal(r.Body)
		return string(str)
	}

	return "{}"
}

// Response represent the structure of real response
type Response struct {
	Status  int                 `json:"status" yaml:"status"`
	Header  map[string][]string `json:"header" yaml:"header"`
	Body    interface{}         `json:"body" yaml:"body"`
	Trailer map[string]string   `json:"trailer" yaml:"trailer"`
	Delay   *ResponseDelay      `json:"delay" yaml:"delay"`
}

func (r *Response) GetBodyString() string {
	if r.Body != nil {
		if reflect.TypeOf(r.Body).Kind() == reflect.String {
			return r.Body.(string)
		}

		if _, ok := r.Body.([]byte); ok {
			return string(r.Body.([]byte))
		}

		str, _ := json.Marshal(r.Body)
		return string(str)
	}

	return "{}"
}

// ResponseDelay represent time delay before server responds.
type ResponseDelay struct {
	Delay  int64 `json:"delay" yaml:"delay"`
	Offset int64 `json:"offset" yaml:"offset"`
}

// GetDelay return random time.Duration with respect to specified time range.
func (d *ResponseDelay) GetDelay() time.Duration {
	offset := d.Offset
	if offset > 0 {
		offset = rand.Int63n(d.Offset)
	}
	return time.Duration(d.Delay+offset) * time.Millisecond
}

// NewDefaultResponse is used to create default response
func NewDefaultResponse(request *Request) *Response {
	var code int
	switch request.Protocol {
	case ProtocolGRPC:
		code = 0
	case ProtocolHTTP:
		code = 1
	}
	return &Response{
		Status:  code,
		Header:  map[string][]string{},
		Trailer: map[string]string{},
		Body:    []byte{},
	}
}
