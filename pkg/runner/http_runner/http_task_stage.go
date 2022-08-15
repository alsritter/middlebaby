package http_runner

import (
	"github.com/alsritter/middlebaby/pkg/apimanager"
	"github.com/alsritter/middlebaby/pkg/runner"
	"github.com/alsritter/middlebaby/pkg/taskserver/task_file"
	"github.com/alsritter/middlebaby/pkg/util/logger"
)

type httpTaskCase struct {
	testCase          task_file.HttpTaskCase
	httpServiceInfo   task_file.HttpTaskInfo
	interfaceOperator task_file.InterfaceOperator

	runner     runner.Runner
	mockCenter apimanager.ApiMockCenter
	log        logger.Logger
}

func newHttpTaskCase(
	interfaceOperator task_file.InterfaceOperator,
	serverInfo task_file.HttpTaskInfo,
	testCase task_file.HttpTaskCase,

	runner runner.Runner,
	mockCenter apimanager.ApiMockCenter,
	log logger.Logger,
) *httpTaskCase {
	return &httpTaskCase{
		testCase:          testCase,
		httpServiceInfo:   serverInfo,
		runner:            runner,
		mockCenter:        mockCenter,
		interfaceOperator: interfaceOperator,
		log:               log,
	}
}

func (r *httpTaskCase) runSetUp() error {
	// run a case level setup first.
	if err := runner.RunSetUp(r.testCase.SetUp, r.mockCenter, r.runner); err != nil {
		r.log.Error(nil, "run a case level setup error: ", err)
		return err
	}

	// then run the interface level setup.
	if err := runner.RunSetUp(r.interfaceOperator.SetUp, r.mockCenter, r.runner); err != nil {
		r.log.Error(nil, "run the interface level setup error: ", err)
		return err
	}
	return nil
}

func (r *httpTaskCase) runTearDown() error {
	// run a case level teardown first.
	if err := runner.RunTearDown(r.testCase.TearDown, r.mockCenter, r.runner); err != nil {
		r.log.Error(nil, "run a case level teardown error: ", err)
		return err
	}

	// then run the interface level teardown.
	if err := runner.RunTearDown(r.interfaceOperator.TearDown, r.mockCenter, r.runner); err != nil {
		r.log.Error(nil, "run a interface level teardown error: ", err)
		return err
	}

	return nil
}

func (r *httpTaskCase) Run() (err error) {
	defer func() {
		if tearDownErr := r.runTearDown(); tearDownErr != nil {
			if err == nil {
				err = tearDownErr
			}
		}
	}()

	if err := r.runSetUp(); err != nil {
		return err
	}

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
	r.log.Debug(nil, "response message: %v %v %v %v \n", responseHeader, responseBody, statusCode, r.testCase.Assert.Response.Data)
	if err := runner.RunHttpAssert(r.testCase.Assert, responseHeader, statusCode, responseBody, r.runner); err != nil {
		return err
	}

	return nil
}
