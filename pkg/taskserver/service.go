package taskserver

import (
	"context"
	"fmt"
	"net/http"

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
}

type TaskService struct {
	// case center
	caseProvider caseprovider.Provider
	// configuration information required by the service.
	cfg *Config

	pluginRegistry pluginregistry.Registry

	log logger.Logger
}

// New return a TaskService
func New(log logger.Logger, cfg *Config,
	caseProvider caseprovider.Provider,
	pluginRegistry pluginregistry.Registry) Provider {

	return &TaskService{
		caseProvider:   caseProvider,
		pluginRegistry: pluginRegistry,
		cfg:            cfg,
		log:            log.NewLogger("task"),
	}
}

func (t *TaskService) Start() error {
	return nil
}

func (t *TaskService) Close() error {
	return nil
}

func (t *TaskService) Run(ctx context.Context, itfName string, caseName string) (err error) {
	var (
		info            = t.caseProvider.GetItfInfoFromItfName(itfName)
		runCase         = t.caseProvider.GetAllCaseFromCaseName(itfName, caseName)
		setupCmds       = t.caseProvider.GetItfSetupCommand(itfName, caseName)
		teardownCmds    = t.caseProvider.GetItfTearDownCommand(itfName, caseName)
		setupCmdType    = make(map[string][]string)
		teardownCmdType = make(map[string][]string)
	)

	for _, c := range setupCmds {
		setupCmdType[c.TypeName] = append(setupCmdType[c.TypeName], c.Commands...)
	}

	for _, c := range teardownCmds {
		teardownCmdType[c.TypeName] = append(setupCmdType[c.TypeName], c.Commands...)
	}

	// before run command
	envs := t.pluginRegistry.EnvPlugins()
	for _, e := range envs {
		if err = e.Run(setupCmdType[e.GetTypeName()]); err != nil {
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
	}()

	if info.Protocol == caseprovider.ProtocolHTTP {
		err = t.httpRequest(runCase)
	} else {
		err = t.grpcRequest(runCase)
	}

	return
}

func (t *TaskService) httpRequest(ct *caseprovider.CaseTask) (err error) {
	// request
	responseHeader, statusCode, responseBody, err := r.runner.Http(
		r.httpServiceInfo.ServiceURL,
		r.httpServiceInfo.ServiceMethod,
		r.testCase.Request.Query,
		r.testCase.Request.Header,
		r.testCase.Request.Data)
	if err != nil {
		return err
	}

	// assert
	t.log.Debug(nil, "response message: %v %v %v %v \n", responseHeader, responseBody, statusCode, ct.Assert.Response.Data)
	if err := t.imposterAssert(ct.Assert, responseHeader, statusCode, responseBody); err != nil {
		return err
	}

	return nil
}

func (t *TaskService) grpcRequest(ct *caseprovider.CaseTask) (err error) {
	return
}

func (t *TaskService) imposterAssert(a caseprovider.ImposterAssert, responseHeader http.Header, statusCode int, responseBody string) error {
	if a.Response.StatusCode != 0 {
		if err := assert.So(t.log, "response status code data assertion", statusCode, a.Response.StatusCode); err != nil {
			return err
		}
	}

	responseKeyVal := make(map[string]string)
	for k := range responseHeader {
		responseKeyVal[k] = responseHeader.Get(k)
	}

	if err := assert.So(t.log, "response header data assertion", responseKeyVal, a.Response.Header); err != nil {
		return err
	}

	if err := assert.So(t.log, "interfaces respond to data assertions", responseBody, a.Response.Data); err != nil {
		return err
	}

	return nil
}
