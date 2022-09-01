package interact

import (
	"encoding/json"
	"math/rand"
	"time"

	"github.com/gogo/protobuf/proto"
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
	Request  Request  `json:"request"`
	Response Response `json:"response"`
}

// Delay returns delay for response that user can specify in imposter config
func (i *ImposterCase) Delay() time.Duration {
	return i.Response.Delay.GetDelay()
}

// Request defines the request structure
type Request struct {
	Protocol Protocol               `json:"protocol"`
	Method   string                 `json:"method"`
	Host     string                 `json:"host"`
	Path     string                 `json:"path"`
	Headers  map[string]interface{} `json:"header"`
	Params   map[string]string      `json:"params"`
	Body     Message                `json:"body"`
}

// Response represent the structure of real response
type Response struct {
	Status  int                 `json:"status"`
	Headers map[string][]string `json:"headers"`
	Body    Message             `json:"body"`
	Trailer map[string]string   `json:"trailer"`
	Delay   ResponseDelay       `json:"delay"`
}

// Message defines a generic message interface
type Message interface {
	proto.Message
	json.Marshaler
	json.Unmarshaler
	proto.Marshaler
	Bytes() []byte
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
		Headers: map[string][]string{},
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
