package task_file

import (
	"github.com/alsritter/middlebaby/pkg/interact"
)

type TestCaseType = string

// TaskCaseTree represents a TaskName and all cases under it.
type TaskCaseTree struct {
	InterfaceName string   // Task Name(Interface Name)
	CaseList      []string // Case Names
}

type SetUp struct {
	Mysql []string
	Redis []string
	HTTP  []interact.HttpImposter
	GRpc  []interact.GRpcImposter
}

type CommonAssert struct {
	Actual   string      // the actual return value of the target.
	Expected interface{} // expect result.
}

type MysqlAssert []CommonAssert
type RedisAssert []CommonAssert

// TearDown the use case completes the post-operation.
type TearDown struct {
	Mysql []string
	Redis []string
}

// InterfaceOperator interface-level operations.
type InterfaceOperator struct {
	SetUp    SetUp    `json:"setup"`
	TearDown TearDown `json:"teardown"`
}
