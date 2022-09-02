package taskserver

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/alsritter/middlebaby/pkg/apimanager"
	"io/ioutil"
	"net/http"
	"net/http/httptrace"
	"net/http/httputil"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/spf13/pflag"

	"github.com/alsritter/middlebaby/pkg/caseprovider"
	"github.com/alsritter/middlebaby/pkg/pluginregistry"
	"github.com/alsritter/middlebaby/pkg/util/assert"
	"github.com/alsritter/middlebaby/pkg/util/logger"
)

type Config struct {
	CaseFiles       []string `yaml:"caseFiles"`
	TaskFileSuffix  string   `yaml:"taskFileSuffix"` // the default test case suffix name. example: ".case.json"
	WatcherCases    bool     `yaml:"watcherCases"`
	MustRunTearDown bool     `yaml:"mustRunTearDown"`
}

func NewConfig() *Config {
	return &Config{}
}

func (c *Config) Validate() error {
	return nil
}

// RegisterFlagsWithPrefix is used to register flags
func (c *Config) RegisterFlagsWithPrefix(prefix string, f *pflag.FlagSet) {}

type Provider interface {
	Start(ctx context.Context, cancelFunc context.CancelFunc, wg *sync.WaitGroup) error
}

type TaskService struct {
	logger.Logger
	cfg            *Config
	caseProvider   caseprovider.Provider
	apiProvider    apimanager.Provider
	pluginRegistry pluginregistry.Registry
}

// New return a TaskService
func New(log logger.Logger, cfg *Config,
	caseProvider caseprovider.Provider,
	apiProvider apimanager.Provider,
	pluginRegistry pluginregistry.Registry,
) Provider {
	return &TaskService{
		cfg:            cfg,
		caseProvider:   caseProvider,
		apiProvider:    apiProvider,
		pluginRegistry: pluginRegistry,
		Logger:         log.NewLogger("task"),
	}
}

func (t *TaskService) Start(ctx context.Context, cancelFunc context.CancelFunc, wg *sync.WaitGroup) error {
	return nil
}

func (t *TaskService) Run(ctx context.Context, itfName string, caseName string) (err error) {
	t.apiProvider.LoadCaseEnv(itfName, caseName)
	var (
		ass             = t.pluginRegistry.AssertPlugins()
		envs            = t.pluginRegistry.EnvPlugins()
		info            = t.caseProvider.GetItfInfoFromItfName(itfName)
		runCase         = t.caseProvider.GetAllCaseFromCaseName(itfName, caseName)
		setupCmds       = t.caseProvider.GetItfSetupCommand(itfName, caseName)
		teardownCmds    = t.caseProvider.GetItfTearDownCommand(itfName, caseName)
		setupCmdType    = make(map[string][]string)
		teardownCmdType = make(map[string][]string)
		assertCmdType   = make(map[string][]caseprovider.CommonAssert)
	)

	for _, c := range setupCmds {
		setupCmdType[c.TypeName] = append(setupCmdType[c.TypeName], c.Commands...)
	}

	for _, c := range teardownCmds {
		teardownCmdType[c.TypeName] = append(setupCmdType[c.TypeName], c.Commands...)
	}

	// before run command

	for _, e := range envs {
		if err := e.Run(setupCmdType[e.GetTypeName()]); err != nil {
			return fmt.Errorf("setup command failed: %v", err)
		}
	}

	// after run command
	defer func() {
		for _, e := range envs {
			if err = e.Run(teardownCmdType[e.GetTypeName()]); err != nil {
				err = fmt.Errorf("teardown command failed: %v", err)
			}
		}
		t.apiProvider.ClearCaseEnv()
	}()

	if err = t.runRequest(info, runCase); err != nil {
		return
	}

	for _, oa := range runCase.Assert.OtherAsserts {
		assertCmdType[oa.TypeName] = append(assertCmdType[oa.TypeName], oa)
	}

	// other assert
	for _, a := range ass {
		if err = a.Assert(assertCmdType[a.GetTypeName()]); err != nil {
			return err
		}
	}
	return
}

func (t *TaskService) runRequest(info *caseprovider.TaskInfo, runCase *caseprovider.CaseTask) error {
	// request assert
	if info.Protocol == caseprovider.ProtocolHTTP {
		return t.httpRequest(info, runCase)
	} else {
		return t.grpcRequest(info, runCase)
	}
}

func (t *TaskService) httpRequest(info *caseprovider.TaskInfo, ct *caseprovider.CaseTask) (err error) {
	// request
	responseHeader, statusCode, responseBody, err := t.http(
		info.ServicePath,
		info.ServiceMethod,
		ct.Request.Query,
		ct.Request.Header,
		ct.Request.Data)
	if err != nil {
		return err
	}

	// assert
	t.Debug(nil, "response message: responseHeader: [%v] responseBody: [%v] statusCode: [%v] Assert.Response: [%v]", responseHeader, responseBody, statusCode, ct.Assert.Response.Data)
	if err := t.imposterAssert(ct.Assert, responseHeader, statusCode, responseBody); err != nil {
		return err
	}

	return nil
}

func (t *TaskService) grpcRequest(info *caseprovider.TaskInfo, ct *caseprovider.CaseTask) (err error) {
	return
}

func (t *TaskService) imposterAssert(a *caseprovider.Assert, responseHeader http.Header, statusCode int, responseBody string) error {
	if a.Response.StatusCode != 0 {
		if err := assert.So(t, "response status code data assertion", statusCode, a.Response.StatusCode); err != nil {
			return err
		}
	}

	responseKeyVal := make(map[string]string)
	for k := range responseHeader {
		responseKeyVal[k] = responseHeader.Get(k)
	}

	if err := assert.So(t, "response header data assertion", responseKeyVal, a.Response.Header); err != nil {
		return err
	}
	if err := assert.So(t, "interfaces respond to data assertions", responseBody, a.Response.Data); err != nil {
		return err
	}
	return nil
}

func (t *TaskService) http(reqUrl, method string, query url.Values, header map[string]string, reqBody interface{}) (http.Header, int, string, error) {
	parseUrl, err := url.Parse(reqUrl)
	if err != nil {
		return nil, 0, "", fmt.Errorf("格式化请求地址错误: %s 错误:%w", reqUrl, err)
	}

	parseQuery := parseUrl.Query()
	for k := range query {
		parseQuery.Add(k, query.Get(k))
	}
	parseUrl.RawQuery = parseQuery.Encode()

	reqBodyStr, err := t.toStrBody(reqBody)
	if err != nil {
		return nil, 0, "", fmt.Errorf("格式化请求数据失败: %s body: %v 错误:%w", reqUrl, reqBody, err)
	}
	reqBodyReader := strings.NewReader(reqBodyStr)

	request, err := http.NewRequest(method, parseUrl.String(), reqBodyReader)
	if err != nil {
		return nil, 0, "", fmt.Errorf("创建请求失败: %s 错误:%w", reqUrl, err)
	}
	for key, val := range header {
		request.Header.Add(key, val)
	}
	client := http.Client{
		Timeout: time.Second * 30,
	}

	trace := &httptrace.ClientTrace{
		GotConn: func(connInfo httptrace.GotConnInfo) {
			fmt.Printf("Got Conn: %+v\n", connInfo)
		},
		DNSDone: func(dnsInfo httptrace.DNSDoneInfo) {
			fmt.Printf("DNS Info: %+v\n", dnsInfo)
		},
	}

	request = request.WithContext(httptrace.WithClientTrace(request.Context(), trace))

	out, _ := httputil.DumpRequest(request, true)
	t.Debug(nil, "%s", out)
	response, err := client.Do(request)
	if err != nil {
		return nil, 0, "", fmt.Errorf("failed to execute the request: [%s] err: [%w]", reqUrl, err)
	}
	resBody := ""
	if response.Body != nil {
		defer response.Body.Close()
		byteBody, err := ioutil.ReadAll(response.Body)
		if err != nil {
			return nil, 0, "", fmt.Errorf("read response failed: [%s] err: [%w]", reqUrl, err)
		}
		resBody = string(byteBody)
	}
	return response.Header, response.StatusCode, resBody, nil
}

func (t *TaskService) toStrBody(reqBody interface{}) (string, error) {
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
