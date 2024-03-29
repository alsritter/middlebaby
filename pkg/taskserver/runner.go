/*
 Copyright (C) 2022 alsritter

 This program is free software: you can redistribute it and/or modify
 it under the terms of the GNU Affero General Public License as
 published by the Free Software Foundation, either version 3 of the
 License, or (at your option) any later version.

 This program is distributed in the hope that it will be useful,
 but WITHOUT ANY WARRANTY; without even the implied warranty of
 MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 GNU Affero General Public License for more details.

 You should have received a copy of the GNU Affero General Public License
 along with this program.  If not, see <https://www.gnu.org/licenses/>.
*/

package taskserver

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptrace"
	"net/http/httputil"
	"net/textproto"
	"net/url"
	"strings"
	"time"

	"github.com/alsritter/middlebaby/pkg/types/mbcase"
	"github.com/alsritter/middlebaby/pkg/util/assert"
	"github.com/alsritter/middlebaby/pkg/util/grpcurl/ext/ggrpcurl"
)

func (t *taskService) Run(ctx context.Context, itfName string, caseName string) (err error) {
	t.apiProvider.LoadCaseEnv(itfName, caseName)
	defer t.apiProvider.ClearCaseEnv()

	var (
		ass              = t.pluginRegistry.AssertPlugins()
		envs             = t.pluginRegistry.EnvPlugins()
		info             = t.caseProvider.GetItfInfoFromItfName(itfName)
		runCase          = t.caseProvider.GetAllCaseFromCaseName(itfName, caseName)
		setupItfCmds     = t.caseProvider.GetItfSetupCommand(itfName)
		setupCaseCmds    = t.caseProvider.GetCaseSetupCommand(itfName, caseName)
		teardownItfCmds  = t.caseProvider.GetItfTearDownCommand(itfName)
		teardownCaseCmds = t.caseProvider.GetCaseTearDownCommand(itfName, caseName)

		setupCmdType    = make(map[string][]string)
		teardownCmdType = make(map[string][]string)
		assertCmdType   = make(map[string][]mbcase.CommonAssert)
	)

	if info == nil || runCase == nil {
		return fmt.Errorf("cannot find case [%s]-[%s]", itfName, caseName)
	}

	for _, c := range setupItfCmds {
		setupCmdType[c.TypeName] = append(setupCmdType[c.TypeName], c.Commands...)
	}

	for _, c := range setupCaseCmds {
		setupCmdType[c.TypeName] = append(setupCmdType[c.TypeName], c.Commands...)
	}

	for _, c := range teardownCaseCmds {
		teardownCmdType[c.TypeName] = append(teardownCmdType[c.TypeName], c.Commands...)
	}

	for _, c := range teardownItfCmds {
		teardownCmdType[c.TypeName] = append(teardownCmdType[c.TypeName], c.Commands...)
	}

	// before run command
	for _, e := range envs {
		if err := e.Run(setupCmdType[e.GetTypeName()]); err != nil {
			return fmt.Errorf("setup command failed: %v", err)
		}
	}

	// after run command
	defer func() {
		if !t.cfg.CloseTearDown {
			for _, e := range envs {
				if tearDownError := e.Run(teardownCmdType[e.GetTypeName()]); tearDownError != nil {
					t.Error(nil, "teardown command failed: %v", tearDownError)
				}
			}
		}
	}()

	ar, err := t.runRequest(info, runCase)
	if err != nil {
		return err
	}

	for _, oa := range runCase.Assert.OtherAsserts {
		assertCmdType[oa.TypeName] = append(assertCmdType[oa.TypeName], oa)
	}

	// other assert
	for _, a := range ass {
		if err = a.Assert(ar, assertCmdType[a.GetTypeName()]); err != nil {
			return err
		}
	}
	return
}

func (t *taskService) runRequest(info *mbcase.TaskInfo, runCase *mbcase.CaseTask) (*mbcase.Response, error) {
	// request assert
	if info.Protocol == mbcase.ProtocolHTTP {
		return t.httpRequest(info, runCase)
	} else {
		return t.grpcRequest(info, runCase)
	}
}

func (t *taskService) httpRequest(info *mbcase.TaskInfo, ct *mbcase.CaseTask) (*mbcase.Response, error) {
	// request
	responseHeader, statusCode, responseBody, err := t.httpClient(
		info.ServicePath,
		info.ServiceMethod,
		ct.Request.Query,
		ct.Request.Header,
		ct.Request.Data)
	if err != nil {
		return nil, err
	}

	// assert
	t.Trace(map[string]interface{}{
		"responseHeader:": responseHeader,
		"responseBody:":   responseBody,
		"statusCode":      statusCode,
		"Assert.Response": ct.Assert.Response,
	}, "response message: ")

	responseKeyVal := make(map[string]string)
	for k := range responseHeader {
		responseKeyVal[k] = responseHeader.Get(k)
	}

	if err := t.imposterAssert(ct.Assert, responseKeyVal, statusCode, responseBody); err != nil {
		return nil, err
	}

	return &mbcase.Response{
		Header:     responseKeyVal,
		Data:       responseBody,
		StatusCode: statusCode,
	}, nil
}

func (t *taskService) grpcRequest(info *mbcase.TaskInfo, ct *mbcase.CaseTask) (*mbcase.Response, error) {
	var addHeaders []string
	for k, v := range ct.Request.Header {
		addHeaders = append(addHeaders, k+":"+v)
	}

	reqBodyStr, err := ct.Request.BodyString()
	if err != nil {
		return nil, err
	}

	dto := ggrpcurl.GGrpCurlDTO{
		Plaintext:     true,
		FormatError:   true,
		EmitDefaults:  true,
		AddHeaders:    addHeaders,
		ImportPaths:   t.protoProvider.GetImportPaths(),
		ProtoFiles:    []string{info.ServiceProtoFile},
		Data:          reqBodyStr,
		ServiceAddr:   t.cfg.TargetServeAdder,
		ServiceMethod: info.ServicePath,
	}

	responseMD, trailerMD, responseBody, _, err := ggrpcurl.NewInvokeGRpc(&dto).Invoke()
	if err != nil {
		t.Error(nil, "grpc request failed, casename: [%s], error:[%v]", ct.Name, err)
	}

	// assert
	t.Trace(map[string]interface{}{
		"responseMD:":     responseMD,
		"trailerMD:":      trailerMD,
		"responseBody:":   responseBody,
		"Assert.Response": ct.Assert.Response,
	}, "response message: ")
	responseKeyVal := make(map[string]string)
	for k := range responseMD {
		responseKeyVal[k] = textproto.MIMEHeader(responseMD).Get(k)
	}

	if err := t.imposterAssert(ct.Assert, responseKeyVal, http.StatusOK, responseBody); err != nil {
		return nil, err
	}

	return &mbcase.Response{
		Header:     responseKeyVal,
		Data:       responseBody,
		StatusCode: http.StatusOK,
	}, nil
}

func (t *taskService) imposterAssert(a *mbcase.Assert, headerKeyVal map[string]string, statusCode int, responseBody string) error {
	if a.Response.StatusCode != 0 {
		if err := assert.So(t, "response status code data assert", statusCode, a.Response.StatusCode); err != nil {
			return err
		}
	}

	if err := assert.So(t, "response header data assert", headerKeyVal, a.Response.Header); err != nil {
		return err
	}
	if err := assert.So(t, "response body data assert", responseBody, a.Response.Data); err != nil {
		return err
	}
	return nil
}

func (t *taskService) httpClient(reqUrl, method string, query url.Values, header map[string]string, reqBody interface{}) (http.Header, int, string, error) {
	parseUrl, err := url.Parse(reqUrl)
	if err != nil {
		return nil, 0, "", fmt.Errorf("format request address error, url:[%s] err:[%v]", reqUrl, err)
	}

	parseQuery := parseUrl.Query()
	for k := range query {
		parseQuery.Add(k, query.Get(k))
	}
	parseUrl.RawQuery = parseQuery.Encode()

	reqBodyStr, err := t.toStrBody(reqBody)
	if err != nil {
		return nil, 0, "", fmt.Errorf("format request body error, url:[%s] body:[%v] error:[%v]", reqUrl, reqBody, err)
	}
	reqBodyReader := strings.NewReader(reqBodyStr)

	request, err := http.NewRequest(method, parseUrl.String(), reqBodyReader)
	if err != nil {
		return nil, 0, "", fmt.Errorf("create request failed, url:[%s] error:[%v]", reqUrl, err)
	}
	for key, val := range header {
		request.Header.Add(key, val)
	}
	client := http.Client{
		Timeout: time.Second * 30,
	}

	trace := &httptrace.ClientTrace{
		GotConn: func(connInfo httptrace.GotConnInfo) {
			t.Trace(nil, "got conn: %+v", connInfo)
		},
	}

	request = request.WithContext(httptrace.WithClientTrace(request.Context(), trace))
	out, _ := httputil.DumpRequestOut(request, true)
	t.Trace(nil, "print the built interface test request \n[%s]", out)

	response, err := client.Do(request)
	if err != nil {
		return nil, 0, "", fmt.Errorf("failed to execute the request: [%s] err: [%v]", reqUrl, err)
	}
	resBody := ""
	if response.Body != nil {
		defer response.Body.Close()
		byteBody, err := ioutil.ReadAll(response.Body)
		if err != nil {
			return nil, 0, "", fmt.Errorf("read response failed: [%s] err: [%v]", reqUrl, err)
		}
		resBody = string(byteBody)
	}
	return response.Header, response.StatusCode, resBody, nil
}
