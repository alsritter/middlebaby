/*
 Copyright (C) 2022 alsritter

 This program is free software: you can redistribute it and/or modify
 it under the terms of the GNU Affero General Public License as
 published by the Free Software Foundation, either version 3 of the
 License, or (at your option) any later version.

 This program is distributed in the hope that it will be useful,
 but WITHOUT ANY WARRANTY; without even the implied warranty of
 MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 GNU Affero General Public License for more details.

 You should have received a copy of the GNU Affero General Public License
 along with this program.  If not, see <https://www.gnu.org/licenses/>.
*/

package interact

import (
	"encoding/json"
	"math/rand"
	"net/textproto"
	"reflect"
	"strings"
	"time"

	"github.com/alsritter/middlebaby/pkg/util"
	"github.com/alsritter/middlebaby/pkg/util/common"
	"github.com/gogo/protobuf/proto"
)

// Protocol defines the protocol of request
type Protocol string

// defines a set of known protocols
const (
	ProtocolHTTP Protocol = "HTTP"
	ProtocolGRPC Protocol = "GRPC"
)

// ImposterMockCase define an imposter structure (a mock case)
type ImposterMockCase struct {
	Id       string   `json:"-"`
	Request  Request  `json:"request" yaml:"request"`
	Response Response `json:"response" yaml:"response"`
}

// Delay returns delay for response that user can specify in imposter config
func (i *ImposterMockCase) Delay() time.Duration {
	return i.Response.Delay.GetDelay()
}

// Request defines the request structure
type Request struct {
	Protocol Protocol            `json:"protocol" yaml:"protocol"`
	Method   string              `json:"method" yaml:"method"`
	Host     string              `json:"host" yaml:"host"`
	Path     string              `json:"path" yaml:"path"`
	Header   map[string][]string `json:"header" yaml:"header"`
	Query    map[string][]string `json:"query" yaml:"query"`
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

func (r *Response) GetByteData() ([]byte, error) {
	strData, ok := r.Body.(string)
	if ok {
		// If it is a string in the form of "@file:filePath",
		// read the contents of the file and send it as a binary stream application/octet-stream
		if strings.HasPrefix(strData, common.StreamFilePrefix) {
			return util.ReadStreamFile(strings.ReplaceAll(strData, common.StreamFilePrefix, ""))
		}

		// If it is a string in the form of "@multiFile:field:fieldName;文件名1:文件1所在地址;......文件名N:文件N所在地址",
		// read the contents of the file and send it as a multipart/form-data
		if strings.HasPrefix(strData, common.MultiFilePrefix) {
			ct := textproto.MIMEHeader(r.Header).Get("Content-Type")
			if ct != "" && ct == "multipart/form-data" {
			} else {
				textproto.MIMEHeader(r.Header).Set("Content-Type", "multipart/form-data")
			}

			return util.ReadMultiFile(strings.ReplaceAll(strData, common.MultiFilePrefix, ""))
		}
		return []byte(strData), nil
	}

	byteData, _ := json.Marshal(r.Body)
	return byteData, nil
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

// Message defines a generic message interface
type Message interface {
	proto.Message
	json.Marshaler
	json.Unmarshaler
	proto.Marshaler
	Bytes() []byte
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
