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

package caseprovider

import (
	"encoding/json"
	"net/url"
	"time"

	"github.com/alsritter/middlebaby/pkg/interact"
)

// Protocol defines the protocol of request
type Protocol string

// defines a set of known protocols
const (
	ProtocolHTTP Protocol = "http"
	ProtocolGRPC Protocol = "grpc"
)

type TaskInfo struct {
	Protocol Protocol `json:"protocol" yaml:"protocol"`
	// ServiceName cannot repeat
	ServiceName string `json:"serviceName" yaml:"serviceName"`
	// if it's grpc interface, it is always POST
	ServiceMethod      string `json:"serviceMethod" yaml:"serviceMethod"` // POST GET PUT
	ServiceDescription string `json:"serviceDescription" yaml:"serviceDescription"`

	// test target
	// http: "/hello"
	// grpc: "/examples.greeter.proto.Greeter/Hello"
	ServicePath string `json:"servicePath" yaml:"servicePath"`

	// if grpc, need protofile path
	ServiceProtoFile string `json:"serviceProtoFile" yaml:"servicePath"`
}

// ItfTask interface level.
type ItfTask struct {
	*TaskInfo
	SetUp    []*Command               `json:"setup" yaml:"setUp"`
	Mocks    []*interact.ImposterCase `json:"mocks" yaml:"mocks"`
	TearDown []*Command               `json:"teardown" yaml:"teardown"`
	Cases    []*CaseTask              `json:"cases" yaml:"cases"`
}

// CaseTask case level
type CaseTask struct {
	Name        string                   `json:"name" yaml:"name"`
	Description string                   `json:"description" yaml:"description"`
	SetUp       []*Command               `json:"setup" yaml:"setup"`
	Mocks       []*interact.ImposterCase `json:"mocks" yaml:"mocks"`
	Request     *CaseRequest             `json:"request" yaml:"request"`
	Assert      *Assert                  `json:"assert" yaml:"assert"`
	TearDown    []*Command               `json:"teardown" yaml:"teardown"`
}

// CaseRequest case request data.
type CaseRequest struct {
	Header map[string]string
	Query  url.Values
	Data   interface{}
}

func (c *CaseRequest) BodyString() (string, error) {
	var reqBodyStr string
	reqBodyStr, ok := c.Data.(string)
	if !ok {
		reqBodyByte, err := json.Marshal(c.Data)
		if err != nil {
			return "", err
		}
		reqBodyStr = string(reqBodyByte)
	}
	return reqBodyStr, nil
}

type Command struct {
	TypeName string   `json:"typeName"` // mysql, redis..
	Commands []string `json:"commands"`
}

type CommonAssert struct {
	TypeName string      `json:"typeName" yaml:"typeName"` // mysql, redis..
	Actual   string      `json:"actual" yaml:"actual"`     // the actual return value of the target.
	Expected interface{} `json:"expected" yaml:"expected"` // the expected return valueresult.
}

func (c *CommonAssert) ExpectedString() string {
	b, _ := json.Marshal(c.Expected)
	return string(b)
}

type Assert struct {
	Response struct {
		Header     map[string]string `json:"header" yaml:"header"`
		Data       interface{}       `json:"data" yaml:"data"`
		StatusCode int               `json:"statusCode" yaml:"statusCode"`
	}
	OtherAsserts []CommonAssert `json:"otherAsserts"`
}

func (a *Assert) ResponseDataString() string {
	b, _ := json.Marshal(a.Response.Data)
	return string(b)
}

// Record some file information
type ItfTaskWithFileInfo struct {
	Dirpath      string    `json:"dirpath" yaml:"dirpath"`
	Filename     string    `json:"filename" yaml:"filename"`
	ModifiedTime time.Time `json:"modifiedTime" yaml:"modifiedTime"`

	*ItfTask
}
