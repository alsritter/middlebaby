package runner

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"

	"github.com/alsritter/middlebaby/pkg/apimanager"
	"github.com/alsritter/middlebaby/pkg/taskserver/task_file"
	"github.com/alsritter/middlebaby/pkg/util/logger"
	"go.elastic.co/apm"
	"google.golang.org/grpc/metadata"
)

var _ Runner = (*defaultRunnerInstance)(nil)

// ITaskRunner grpc or http runner interface.
type ITaskRunner interface {
	// Run execution test case.
	Run(caseName string, mockCenter apimanager.ApiMockCenter, runner Runner) error
	// GetTaskCaseTree Get All Task and the Task's Cases
	GetTaskCaseTree() []*task_file.TaskCaseTree
}

// Runner runner group
type Runner interface {
	// MySQL exec SQL.
	MySQL(sql string) ([]map[string]interface{}, error)
	// Redis Start the Redis command.
	Redis(cmd string) (interface{}, error)
	// GRpc request.
	GRpc(serviceProtoFile, serviceMethod, appServeAddr string, protoPaths []string, reqHeader map[string]string, reqBody interface{}) (md metadata.MD, body interface{}, err error)
	// Http request.
	Http(url, method string, query url.Values, header map[string]string, body interface{}) (http.Header, int, string, error)
	// Clone a Runner.
	Clone() Runner
	// RunID The current Runner uniquely id.
	RunID() string
}

type RedisRunner interface {
	Run(cmd string) (result interface{}, err error)
}

type MysqlRunner interface {
	Run(sql string) (result []map[string]interface{}, err error)
}

type defaultRunnerInstance struct {
	mysqlRunner  MysqlRunner
	redisRunner  RedisRunner
	traceContext apm.TraceContext // generate a trace id.
	log          logger.Logger
}

// NewRunner return a runner.
func NewRunner(mysqlRunner MysqlRunner, redisRunner RedisRunner, log logger.Logger) (Runner, error) {
	return &defaultRunnerInstance{
		mysqlRunner: mysqlRunner,
		redisRunner: redisRunner,
		log:         log,
	}, nil
}

func (c *defaultRunnerInstance) MySQL(sql string) (result []map[string]interface{}, err error) {
	return c.mysqlRunner.Run(sql)
}

func (c *defaultRunnerInstance) Redis(cmd string) (res interface{}, err error) {
	return c.redisRunner.Run(cmd)
}

// GRpc TODO: do something....
func (c *defaultRunnerInstance) GRpc(serviceProtoFile, serviceMethod, appServeAddr string, protoPaths []string, reqHeader map[string]string, reqBody interface{}) (md metadata.MD, body interface{}, err error) {
	return
}

// Http request.
func (c *defaultRunnerInstance) Http(reqUrl, method string, query url.Values, header map[string]string, body interface{}) (http.Header, int, string, error) {
	parseUrl, err := url.Parse(reqUrl)
	if err != nil {
		return nil, 0, "", fmt.Errorf("error formatting request address: %s error: %w", reqUrl, err)
	}

	parseQuery := parseUrl.Query()
	for k := range query {
		parseQuery.Add(k, query.Get(k))
	}
	parseUrl.RawQuery = parseQuery.Encode()
	reqBodyStr, err := c.toStrBody(body)
	if err != nil {
		return nil, 0, "", fmt.Errorf("error formatting %s request data body: %v error:%w", reqUrl, body, err)
	}

	reqBodyReader := strings.NewReader(reqBodyStr)
	request, err := http.NewRequest(method, parseUrl.String(), reqBodyReader)
	if err != nil {
		return nil, 0, "", fmt.Errorf("create request %s error: %w", reqUrl, err)
	}

	for key, val := range header {
		request.Header.Add(key, val)
	}
	client := http.Client{
		Timeout: time.Second * 30,
	}

	out, _ := httputil.DumpRequest(request, true)
	c.log.Debug(nil, string(out))
	response, err := client.Do(request)
	if err != nil {
		return nil, 0, "", fmt.Errorf("request execution failed, error %w by %s", err, reqUrl)
	}

	resBody := ""
	if response.Body != nil {
		defer response.Body.Close()
		byteBody, err := ioutil.ReadAll(response.Body)
		if err != nil {
			return nil, 0, "", fmt.Errorf("read response failure, error %w by %s", err, reqUrl)
		}
		resBody = string(byteBody)
	}
	return response.Header, response.StatusCode, resBody, nil
}

// Clone generate a new trace id.
func (c *defaultRunnerInstance) Clone() Runner {
	traceContext := apm.DefaultTracer.StartTransaction("middlebaby", "test").TraceContext()
	cn := *c
	cn.traceContext = traceContext
	return &cn
}

// RunID The current Runner uniquely id.
func (c *defaultRunnerInstance) RunID() string {
	// e.g., 00000000000000000000000000000000
	return c.traceContext.Trace.String()
}

// request data to json string.
func (c *defaultRunnerInstance) toStrBody(reqBody interface{}) (string, error) {
	var reqBodyStr string
	reqBodyStr, ok := reqBody.(string)
	if !ok {
		reqBodyByte, err := json.Marshal(reqBody)
		if err != nil {
			return "", err
		}
		reqBodyStr = string(reqBodyByte)
	}
	return reqBodyStr, nil
}
