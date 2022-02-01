package task

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"

	"alsritter.icu/middlebaby/internal/log"
	"go.elastic.co/apm"
	"google.golang.org/grpc/metadata"
)

var _ (Runner) = (*runner)(nil)

// runner group
type Runner interface {
	// exec SQL.
	MySQL(sql string) ([]map[string]interface{}, error)
	// Run the Redis command.
	Redis(cmd string) (interface{}, error)
	// GRpc request.
	GRpc(serviceProtoFile, serviceMethod, appServeAddr string, protoPaths []string, reqHeader map[string]string, reqBody interface{}) (md metadata.MD, body interface{}, err error)
	// Http request.
	Http(url, method string, query url.Values, header map[string]string, body interface{}) (http.Header, int, string, error)
	// Clone a Runner.
	Clone() Runner
	// The current Runner uniquely id.
	RunID() string
}

type RedisRunner interface {
	Run(cmd string) (result interface{}, err error)
}

type MysqlRunner interface {
	Run(sql string) (result []map[string]interface{}, err error)
}

type runner struct {
	mysqlRunner  MysqlRunner
	redisRunner  RedisRunner
	traceContext apm.TraceContext // generate a trace id.
}

// return a runner.
func NewRunner(mysqlRunner MysqlRunner, redisRunner RedisRunner) (Runner, error) {
	return &runner{
		mysqlRunner: mysqlRunner,
		redisRunner: redisRunner,
	}, nil
}

func (c *runner) MySQL(sql string) (result []map[string]interface{}, err error) {
	return c.mysqlRunner.Run(sql)
}

func (c *runner) Redis(cmd string) (res interface{}, err error) {
	return c.redisRunner.Run(cmd)
}

// TODO: do something....
func (c *runner) GRpc(serviceProtoFile, serviceMethod, appServeAddr string, protoPaths []string, reqHeader map[string]string, reqBody interface{}) (md metadata.MD, body interface{}, err error) {
	return
}

// Http request.
func (c *runner) Http(reqUrl, method string, query url.Values, header map[string]string, body interface{}) (http.Header, int, string, error) {
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
	log.Debug(out)
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

// generate a new trace id.
func (c *runner) Clone() Runner {
	traceContext := apm.DefaultTracer.StartTransaction("middlebaby", "test").TraceContext()
	cn := *c
	cn.traceContext = traceContext
	return &cn
}

// The current Runner uniquely id.
func (c *runner) RunID() string {
	return c.traceContext.Trace.String()
}

// request data to json string.
func (c *runner) toStrBody(reqBody interface{}) (string, error) {
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
