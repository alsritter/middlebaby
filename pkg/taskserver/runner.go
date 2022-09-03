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

	"github.com/alsritter/middlebaby/pkg/caseprovider"
	"github.com/alsritter/middlebaby/pkg/util/assert"
	"github.com/alsritter/middlebaby/pkg/util/grpcurl/ext/ggrpcurl"
)

func (t *taskService) Run(ctx context.Context, itfName string, caseName string) (err error) {
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
		if !t.cfg.CloseTearDown {
			for _, e := range envs {
				if err = e.Run(teardownCmdType[e.GetTypeName()]); err != nil {
					err = fmt.Errorf("teardown command failed: %v", err)
				}
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

func (t *taskService) runRequest(info *caseprovider.TaskInfo, runCase *caseprovider.CaseTask) error {
	// request assert
	if info.Protocol == caseprovider.ProtocolHTTP {
		return t.httpRequest(info, runCase)
	} else {
		return t.grpcRequest(info, runCase)
	}
}

func (t *taskService) httpRequest(info *caseprovider.TaskInfo, ct *caseprovider.CaseTask) (err error) {
	// request
	responseHeader, statusCode, responseBody, err := t.httpClient(
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

	responseKeyVal := make(map[string]string)
	for k := range responseHeader {
		responseKeyVal[k] = responseHeader.Get(k)
	}

	if err := t.imposterAssert(ct.Assert, responseKeyVal, statusCode, responseBody); err != nil {
		return err
	}

	return nil
}

func (t *taskService) grpcRequest(info *caseprovider.TaskInfo, ct *caseprovider.CaseTask) (err error) {
	var addHeaders []string
	for k, v := range ct.Request.Header {
		addHeaders = append(addHeaders, k+":"+v)
	}

	reqBodyStr, err := ct.Request.BodyString()
	if err != nil {
		return err
	}

	dto := ggrpcurl.GGrpCurlDTO{
		Plaintext:     true,
		FormatError:   true,
		EmitDefaults:  true,
		AddHeaders:    addHeaders,
		ImportPaths:   t.protoProvider.GetImportPaths(),
		ProtoFiles:    []string{info.ServiceProtoFile},
		Data:          reqBodyStr,
		ServiceAddr:   info.ServicePath,
		ServiceMethod: info.ServiceMethod,
	}

	responseMD, responseBody, _, err := ggrpcurl.NewInvokeGRpc(&dto).Invoke()
	if err != nil {
		t.Error(nil, "grpc request failed, casename: [%s], error:[%w]", ct.Name, err)
	}

	// assert
	t.Debug(nil, "response message: responseMD: [%v] responseBody: [%v] Assert.Response: [%v]", responseMD, responseBody, ct.Assert.Response.Data)
	responseKeyVal := make(map[string]string)
	for k := range responseMD {
		responseKeyVal[k] = textproto.MIMEHeader(responseMD).Get(k)
	}

	if err := t.imposterAssert(ct.Assert, responseKeyVal, http.StatusOK, responseBody); err != nil {
		return err
	}

	return nil
}

func (t *taskService) imposterAssert(a *caseprovider.Assert, responseKeyVal map[string]string, statusCode int, responseBody string) error {
	if a.Response.StatusCode != 0 {
		if err := assert.So(t, "response status code data assertion", statusCode, a.Response.StatusCode); err != nil {
			return err
		}
	}

	if err := assert.So(t, "response header data assertion", responseKeyVal, a.Response.Header); err != nil {
		return err
	}
	if err := assert.So(t, "interfaces respond to data assertions", responseBody, a.Response.Data); err != nil {
		return err
	}
	return nil
}

func (t *taskService) httpClient(reqUrl, method string, query url.Values, header map[string]string, reqBody interface{}) (http.Header, int, string, error) {
	parseUrl, err := url.Parse(reqUrl)
	if err != nil {
		return nil, 0, "", fmt.Errorf("format request address error, url:[%s] err:[%w]", reqUrl, err)
	}

	parseQuery := parseUrl.Query()
	for k := range query {
		parseQuery.Add(k, query.Get(k))
	}
	parseUrl.RawQuery = parseQuery.Encode()

	reqBodyStr, err := t.toStrBody(reqBody)
	if err != nil {
		return nil, 0, "", fmt.Errorf("format request body error, url:[%s] body:[%v] error:[%w]", reqUrl, reqBody, err)
	}
	reqBodyReader := strings.NewReader(reqBodyStr)

	request, err := http.NewRequest(method, parseUrl.String(), reqBodyReader)
	if err != nil {
		return nil, 0, "", fmt.Errorf("create request failed, url:[%s] error:[%w]", reqUrl, err)
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
