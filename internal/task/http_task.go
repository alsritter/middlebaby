package task

import (
	"alsritter.icu/middlebaby/internal/file/task_file"
	"alsritter.icu/middlebaby/internal/log"
	"alsritter.icu/middlebaby/internal/proxy"
	"alsritter.icu/middlebaby/internal/startup/plugin"
)

// an interface under test is tested.
type TargetTask struct {
	// name of the interface to be tested.
	TargetInterfaceName string
	// all test cases for the current interface
	TaskCaseList []*TaskCase
}

type TaskCase struct {
	CaseName string
}

type HttpTaskCase struct {
	testCase          task_file.HttpTaskCase
	httpServiceInfo   task_file.HttpTaskInfo
	runner            Runner
	mockCenter        proxy.MockCenter
	env               plugin.Env
	interfaceOperator task_file.InterfaceOperator
}

func (r *HttpTaskCase) runSetUp() error {
	// run a case level setup first.
	if err := RunSetUp(r.testCase.SetUp, r.mockCenter, r.runner); err != nil {
		log.Error("run a case level setup error: ", err)
		return err
	}

	// then run the interface level setup.
	if err := RunSetUp(r.interfaceOperator.SetUp, r.mockCenter, r.runner); err != nil {
		log.Error("run the interface level setup error: ", err)
		return err
	}
	return nil
}

func (r *HttpTaskCase) runTearDown() error {
	// run a case level teardown first.
	if err := RunTearDown(r.testCase.TearDown, r.mockCenter, r.runner); err != nil {
		log.Error("run a case level teardown error: ", err)
		return err
	}

	// then run the interface level teardown.
	if err := RunTearDown(r.interfaceOperator.TearDown, r.mockCenter, r.runner); err != nil {
		log.Error("run a interface level teardown error: ", err)
		return err
	}

	return nil
}

func (r *HttpTaskCase) Run() (err error) {
	defer func() {
		if r.env.GetMustRunTearDown() || err == nil {
			if tearDownErr := r.runTearDown(); tearDownErr != nil {
				if err == nil {
					err = tearDownErr
				}
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
	if err := RunHttpAssert(r.testCase.Assert, responseHeader, statusCode, responseBody, r.runner); err != nil {
		return err
	}

	return nil
}
