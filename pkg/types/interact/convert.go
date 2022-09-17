package interact

import (
	"bytes"
	"compress/gzip"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/alsritter/middlebaby/pkg/util/grpcurl/ext/ggrpcurl"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func HttpConverter(req *http.Request, resp *http.Response) (*ImposterMockCase, error) {
	var (
		outresp Response
		outreq  Request
	)

	if req == nil {
		return nil, errors.New("req cannot be nil")
	} else {
		body, err := ioutil.ReadAll(req.Body)
		if err != nil {
			return nil, err
		}
		req.Body = ioutil.NopCloser(bytes.NewReader(body))
		outreq = Request{
			Protocol: ProtocolHTTP,
			Method:   req.Method,
			Host:     req.Host,
			Path:     req.URL.Path,
			Header:   req.Header,
			Query:    req.URL.Query(),
			Body:     string(body),
		}
	}

	if resp == nil {
		outresp = Response{}
	} else {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		resp.Body = ioutil.NopCloser(bytes.NewReader(body))

		for _, curEnc := range strings.Split(req.Header.Get("Accept-Encoding"), ",") {
			curEnc = strings.TrimSpace(curEnc)
			if curEnc == "gzip" {
				gr, err := gzip.NewReader(ioutil.NopCloser(bytes.NewReader(body))) //初始化gzip reader
				if err != nil {
					return nil, fmt.Errorf("resp body gzip parse failure: [%v]", err)
				}
				body, _ = ioutil.ReadAll(gr)
				break
			}
		}

		outresp = Response{
			Status:  resp.StatusCode,
			Header:  resp.Header,
			Body:    string(body),
			Trailer: resp.Trailer,
			Delay:   &ResponseDelay{},
		}
	}

	return &ImposterMockCase{
		Request:  outreq,
		Response: outresp,
	}, nil
}

func GRPCConvert(reqMD metadata.MD, dto ggrpcurl.GGrpCurlDTO,
	respMD metadata.MD, respTrailer metadata.MD,
	respBody string, s *status.Status) *ImposterMockCase {
	var (
		outresp = Response{
			Status:  int(s.Code()),
			Header:  respMD,
			Body:    respBody,
			Trailer: respTrailer,
			Delay:   &ResponseDelay{},
		}

		outreq = Request{
			Protocol: ProtocolHTTP,
			Method:   http.MethodPost,
			Host:     dto.ServiceAddr,
			Path:     dto.ServiceMethod,
			Header:   reqMD,
			Query:    nil,
			Body:     dto.Data,
		}
	)

	return &ImposterMockCase{
		Request:  outreq,
		Response: outresp,
	}
}
