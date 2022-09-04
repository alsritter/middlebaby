package taskserver

import (
	"encoding/json"

	"github.com/alsritter/middlebaby/pkg/caseprovider"
	"github.com/alsritter/middlebaby/pkg/interact"
	"github.com/alsritter/middlebaby/pkg/util"
	taskproto "github.com/alsritter/middlebaby/proto/task"
)

func (t *taskService) toStrBody(reqBody interface{}) (string, error) {
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

func (t *taskService) toGetAllTaskCasesReply(all []*caseprovider.InterfaceTask) *taskproto.GetAllTaskCasesReply {
	is := make([]*taskproto.InterfaceTask, 0, len(all))
	for _, v := range all {
		is = append(is, t.toProtoInterfaceTask(v))
	}
	return &taskproto.GetAllTaskCasesReply{
		Itfs: is,
	}
}

func (t *taskService) toProtoInterfaceTask(c *caseprovider.InterfaceTask) *taskproto.InterfaceTask {

	return &taskproto.InterfaceTask{
		Protocol:           string(c.Protocol),
		ServiceName:        c.ServiceName,
		ServiceMethod:      c.ServiceMethod,
		ServiceDescription: c.ServiceDescription,
		ServicePath:        c.ServicePath,
		SetUp:              t.toProtoCommandList(c.SetUp),
		Teardown:           t.toProtoCommandList(c.TearDown),
		Mocks:              t.toProtoMockList(c.Mocks),
		Cases:              t.toProtoCaseList(c.Cases),
	}
}

func (t *taskService) toProtoCommandList(cs []*caseprovider.Command) []*taskproto.Command {
	dd := make([]*taskproto.Command, 0, len(cs))
	for _, v := range cs {
		dd = append(dd, t.toProtoCommand(v))
	}
	return dd
}

func (t *taskService) toProtoMockList(ms []*interact.ImposterCase) []*taskproto.ImposterCase {
	mk := make([]*taskproto.ImposterCase, 0, len(ms))
	for _, v := range ms {
		mk = append(mk, t.toProtoImposterCase(v))
	}
	return mk
}

func (t *taskService) toProtoCaseList(cs []*caseprovider.CaseTask) []*taskproto.TaskCase {
	dd := make([]*taskproto.TaskCase, 0, len(cs))
	for _, v := range cs {
		dd = append(dd, t.toProtoCase(v))
	}
	return dd
}

func (t *taskService) toProtoCase(c *caseprovider.CaseTask) *taskproto.TaskCase {
	return &taskproto.TaskCase{
		Name:        c.Name,
		Description: c.Description,
		Assert: &taskproto.Assert{
			Response: &taskproto.AssertResponse{
				Header:     c.Assert.Response.Header,
				Data:       c.Assert.ResponseDataString(),
				StatusCode: int32(c.Assert.Response.StatusCode),
			},
			OtherAsserts: t.toProtoCommonAssert(c.Assert.OtherAsserts),
		},
		SetUp:    t.toProtoCommandList(c.SetUp),
		Teardown: t.toProtoCommandList(c.TearDown),
		Mocks:    t.toProtoMockList(c.Mocks),
	}
}

func (t *taskService) toProtoCommonAssert(as []caseprovider.CommonAssert) []*taskproto.CommonAssert {
	dd := make([]*taskproto.CommonAssert, 0, len(as))
	for _, v := range as {
		dd = append(dd, &taskproto.CommonAssert{
			TypeName: v.TypeName,
			Actual:   v.Actual,
			Expected: v.ExpectedString(),
		})
	}
	return dd
}

func (t *taskService) toProtoCommand(c *caseprovider.Command) *taskproto.Command {
	return &taskproto.Command{
		TypeName: c.TypeName,
		Commands: c.Commands,
	}
}

func (t *taskService) toProtoImposterCase(i *interact.ImposterCase) *taskproto.ImposterCase {
	return &taskproto.ImposterCase{
		Request: &taskproto.Request{
			Protocol: string(i.Request.Protocol),
			Method:   i.Request.Method,
			Host:     i.Request.Host,
			Path:     i.Request.Path,
			Header:   util.SliceMapToStringMap(i.Request.Header),
			Params:   i.Request.Params,
			Body:     i.Request.GetBodyString(),
		},
		Response: &taskproto.Response{
			Status:  int32(i.Response.Status),
			Header:  util.SliceMapToStringMap(i.Response.Header),
			Trailer: i.Response.Trailer,
			Body:    i.Response.GetBodyString(),
		},
	}
}
