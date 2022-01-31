package task_file

import "net/url"

type GRpcTask struct {
	*GRpcTaskInfo
	Cases []*HttpTaskCase `json:"cases"`
	*InterfaceOperator
}

type GRpcTaskInfo struct {
	ServiceName        string `json:"serviceName"`
	ServiceDescription string `json:"ServiceDescription"`
	ServiceProtoFile   string `json:"serviceProtoFile"`
	ServiceMethod      string `json:"serviceMethod"`
}

type GRpcTaskCase struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	SetUp       SetUp           `json:"setup"`
	Request     GRpcCaseRequest `json:"request"`
	Assert      GRpcAssert      `json:"assert"`
	TearDown    TearDown        `json:"teardown"`
}

// case request data.
type GRpcCaseRequest struct {
	Header map[string]string
	Query  url.Values
	Data   interface{}
}

// assertions data.
type GRpcAssert struct {
	Response struct {
		Header     map[string]string
		Data       interface{}
		StatusCode int
	}

	Mysql MysqlAssert `json:"mysql"`
	Redis RedisAssert `json:"redis"`
}
