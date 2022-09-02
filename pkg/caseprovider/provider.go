package caseprovider

import (
	"net/url"

	"github.com/alsritter/middlebaby/pkg/interact"
)

type Provider interface {
	// GetAllCaseFromItfName Get all cases form the interface serviceName.
	GetAllCaseFromItfName(serviceName string) []*CaseTask
	GetAllCaseFromCaseName(serviceName, caseName string) *CaseTask

	GetItfInfoFromItfName(serviceName string) *TaskInfo
	// GetAllItfInfo Get all interface info.
	GetAllItfInfo() []*TaskInfo

	// GetItfSetupCommand Get the Setup Commands of a type under the interface.
	GetItfSetupCommand(serviceName, typeName string) []*Command
	// GetItfTearDownCommand Get the TearDown Commands of a type under the interface.
	GetItfTearDownCommand(serviceName, typeName string) []*Command

	GetMockCasesFromGlobals() []*interact.ImposterCase
	GetMockCasesFromItf(serviceName string) []*interact.ImposterCase
	GetMockCasesFromCase(serviceName, caseName string) []*interact.ImposterCase
}

// InterfaceTask interface level.
type InterfaceTask struct {
	*TaskInfo
	SetUp    []*Command               `json:"setup"`
	Mocks    []*interact.ImposterCase `json:"mocks"`
	TearDown []*Command               `json:"teardown"`
	Cases    []*CaseTask              `json:"cases"`
}

// CaseTask case level
type CaseTask struct {
	Name        string                   `json:"name"`
	Description string                   `json:"description"`
	SetUp       []*Command               `json:"setup"`
	Mocks       []*interact.ImposterCase `json:"mocks"`
	Request     *CaseRequest             `json:"request"`
	Assert      *Assert                  `json:"assertprovid"`
	TearDown    []*Command               `json:"teardown"`
}

// CaseRequest case request data.
type CaseRequest struct {
	Header map[string]string
	Query  url.Values
	Data   interface{}
}

type Command struct {
	TypeName string   `json:"typeName"` // mysql, redis..
	Commands []string `json:"commands"`
}

type CommonAssert struct {
	TypeName string      `json:"typeName"` // mysql, redis..
	Actual   string      // the actual return value of the target.
	Expected interface{} // expect result.
}

type Assert struct {
	Response struct {
		Header     map[string]string
		Data       interface{}
		StatusCode int
	}

	OtherAsserts []CommonAssert `json:"otherAsserts"`
}

// Protocol defines the protocol of request
type Protocol string

// defines a set of known protocols
const (
	ProtocolHTTP Protocol = "HTTP"
	ProtocolGRPC Protocol = "GRPC"
)

type TaskInfo struct {
	Protocol Protocol `json:"protocol"`
	// ServiceName cannot repeat
	ServiceName string `json:"serviceName"`
	// if it's grpc interface, it is always POST
	ServiceMethod      string `json:"serviceMethod"` // POST GET PUT
	ServiceDescription string `json:"serviceDescription"`

	// test target
	// http: "/hello"
	// grpc: "/examples.greeter.proto.Greeter/Hello"
	ServicePath string `json:"servicePath"`
}
