package task_file

import (
	"alsritter.icu/middlebaby/internal/file/common"
)

type SetUp struct {
	Mysql []string
	Redis []string
	HTTP  []common.HttpImposter
	GRpc  []common.GRpcImposter
}

type CommonAssert struct {
	Actual   string      // the actual return value of the target.
	Expected interface{} // expect result.
}

type MysqlAssert []CommonAssert
type RedisAssert []CommonAssert

// the use case completes the post-operation.
type TearDown struct {
	Mysql []string
	Redis []string
}

// interface-level operations.
type InterfaceOperator struct {
	SetUp    SetUp    `json:"setup"`
	TearDown TearDown `json:"teardown"`
}
