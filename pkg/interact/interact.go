package interact

import (
	"encoding/json"
	"math/rand"
	"time"

	"github.com/gogo/protobuf/proto"
)

// Imposter define an imposter structure
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

// Delay return random time.Duration with respect to specified time range.
func (d *ResponseDelay) GetDelay() time.Duration {
	offset := d.Offset
	if offset > 0 {
		offset = rand.Int63n(d.Offset)
	}
	return time.Duration(d.Delay+offset) * time.Millisecond
}

// TODO: fill in the details.
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

// =========================================================================================

// Protocol defines the protocol of request
type Protocol string

// defines a set of known protocols
const (
	ProtocolHTTP Protocol = "HTTP"
	ProtocolGRPC Protocol = "GRPC"
)

// Message defines a generic message interface
type Message interface {
	proto.Message
	json.Marshaler
	json.Unmarshaler
	proto.Marshaler
	Bytes() []byte
}

// Request defines the request structure
type Request struct {
	Protocol Protocol          `json:"protocol"`
	Method   string            `json:"method"`
	Host     string            `json:"host"`
	Path     string            `json:"path"`
	Header   map[string]string `json:"header"`
	Body     Message           `json:"body"`
}

// Response defines the response structure
type Response struct {
	Code    uint32            `json:"code"`
	Header  map[string]string `json:"header"`
	Body    Message           `json:"body"`
	Trailer map[string]string `json:"trailer"`
}

// NewDefaultResponse is used to create default response
func NewDefaultResponse(request *Request) *Response {
	var code uint32
	switch request.Protocol {
	case ProtocolGRPC:
		code = 0
	case ProtocolHTTP:
		code = 1
	}
	return &Response{
		Code:    code,
		Header:  map[string]string{},
		Trailer: map[string]string{},
		Body:    NewBytesMessage(nil),
	}
}

// BytesMessage is the simple implement of Message
type BytesMessage struct {
	data []byte
}

// NewBytesMessage is used to init BytesMessage
func NewBytesMessage(data []byte) Message {
	return &BytesMessage{
		data: data,
	}
}

// Reset implements the proto.Message interface
func (b *BytesMessage) Reset() {}

// String implements the proto.Message interface
func (b *BytesMessage) String() string {
	return string(b.data)
}

// ProtoMessage implements the proto.Message interface
func (b *BytesMessage) ProtoMessage() {}

// Marshal implements the proto.Marshaler interface
func (b *BytesMessage) Marshal() ([]byte, error) {
	return b.data, nil
}

// UnmarshalJSON implements the json.UnmarshalJSON interface
func (b *BytesMessage) UnmarshalJSON(data []byte) error {
	b.data = data
	return nil
}

// MarshalJSON implements the json.Marshaler interface
func (b *BytesMessage) MarshalJSON() ([]byte, error) {
	if len(b.data) == 0 {
		return []byte(`null`), nil
	}
	return b.data, nil
}

// Bytes is used to return native data
func (b *BytesMessage) Bytes() []byte {
	return b.data
}
