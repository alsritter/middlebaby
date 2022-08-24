package runner

import (
	"fmt"
	"net/http"

	"github.com/alsritter/middlebaby/pkg/apimanager"
	"github.com/alsritter/middlebaby/pkg/taskserver/task_file"
	"github.com/alsritter/middlebaby/pkg/util/assert"
)

// RunSetUp setup run
func RunSetUp(s task_file.SetUp, mockCenter apimanager.MockCaseCenter, runner Runner) error {
	for _, sql := range s.Mysql {
		if _, err := runner.MySQL(sql); err != nil {
			return fmt.Errorf("execution SetUp.Mysql error: %w", err)
		}
	}
	for _, cmd := range s.Redis {
		if _, err := runner.Redis(cmd); err != nil {
			return fmt.Errorf("execution SetUp.Redis error: %w", err)
		}
	}
	// get current all HTTP mock Ids.
	var httpIdList []string
	for _, httpMock := range s.HTTP {
		if httpMock.Id != "" {
			httpIdList = append(httpIdList, httpMock.Id)
		}
	}

	// empty all HTTP mocks first to avoid id duplication.
	mockCenter.UnloadHttpByIdList(runner.RunID(), httpIdList)
	for _, httpMock := range s.HTTP {
		// assigning a new ID.
		if err := mockCenter.AddHttp(runner.RunID(), httpMock); err != nil {
			return fmt.Errorf("execution SetUp.Http error: %w", err)
		}
	}

	// same as above.
	var grpcIdList []string
	for _, grpcMock := range s.GRpc {
		if grpcMock.Id != "" {
			grpcIdList = append(grpcIdList, grpcMock.Id)
		}
	}
	mockCenter.UnloadGRpcByIdList(runner.RunID(), grpcIdList)
	for _, gRpcMock := range s.GRpc {
		if err := mockCenter.AddGRpc(runner.RunID(), gRpcMock); err != nil {
			return fmt.Errorf("execution SetUp.GRpc error: %w", err)
		}
	}
	return nil
}

// RunMySQLAssert run mysql assert.
func RunMySQLAssert(m task_file.MysqlAssert, runner Runner) error {
	for _, sqlAssert := range m {
		if result, err := runner.MySQL(sqlAssert.Actual); err != nil {
			return err
		} else if len(result) <= 0 {
			return fmt.Errorf("no result is found: %s", sqlAssert.Actual)

			// this result[0] returns a map
		} else if err := assert.So("MySQL data assert", result[0], sqlAssert.Expected); err != nil {
			return err
		}
	}
	return nil
}

// RunRedisAssert run redis assert.
func RunRedisAssert(r task_file.RedisAssert, runner Runner) error {
	for _, redisAssert := range r {
		if result, err := runner.Redis(redisAssert.Actual); err != nil {
			return err
		} else if err := assert.So("Redis data assert", result, redisAssert.Expected); err != nil {
			return err
		}
	}
	return nil
}

// RunHttpAssert run http assert.
func RunHttpAssert(a task_file.HttpAssert, responseHeader http.Header, statusCode int, responseBody string, runner Runner) error {
	if a.Response.StatusCode != 0 {
		if err := assert.So("response status code data assertion", statusCode, a.Response.StatusCode); err != nil {
			return err
		}
	}

	responseKeyVal := make(map[string]string)
	for k := range responseHeader {
		responseKeyVal[k] = responseHeader.Get(k)
	}

	if err := assert.So("response header data assertion", responseKeyVal, a.Response.Header); err != nil {
		return err
	}

	if err := assert.So("interfaces respond to data assertions", responseBody, a.Response.Data); err != nil {
		return err
	}

	if err := RunMySQLAssert(a.Mysql, runner); err != nil {
		return err
	}

	if err := RunRedisAssert(a.Redis, runner); err != nil {
		return err
	}

	return nil
}

// RunTearDown run tearDown.
func RunTearDown(t task_file.TearDown, mockCenter apimanager.MockCaseCenter, runner Runner) error {
	// when the task server is complete, empty the mock for the current case.
	mockCenter.UnLoadHttp(runner.RunID())
	mockCenter.UnLoadGRpc(runner.RunID())

	for _, sql := range t.Mysql {
		if _, err := runner.MySQL(sql); err != nil {
			return fmt.Errorf("execution tearDown.Mysql error: %w", err)
		}
	}

	for _, cmd := range t.Redis {
		if _, err := runner.Redis(cmd); err != nil {
			return fmt.Errorf("execution tearDown.Redis error: %w", err)
		}
	}

	return nil
}
