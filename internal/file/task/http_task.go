package task_file

import "net/url"

type HttpTask struct {
	*HttpServiceInfo
	Cases []*HttpTaskCase `json:"cases"`
	*InterfaceOperator
}

type HttpServiceInfo struct {
	ServiceName        string `json:"serviceName"`
	ServiceDescription string `json:"ServiceDescription"`
	ServiceURL         string `json:"serviceURL"`
}

type HttpTaskCase struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	SetUp       SetUp           `json:"setup"`
	Request     HttpCaseRequest `json:"request"`
	Assert      HttpAssert      `json:"assert"`
	TearDown    TearDown        `json:"teardown"`
}

// case request data.
type HttpCaseRequest struct {
	Header map[string]string
	Query  url.Values
	Data   interface{}
}

// assertions data.
type HttpAssert struct {
	Response struct {
		Header     map[string]string
		Data       interface{}
		StatusCode int
	}

	Mysql MysqlAssert `json:"mysql"`
	Redis RedisAssert `json:"redis"`
}
