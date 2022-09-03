// Command grpcurl makes gRPC requests (a la cURL, but HTTP/2). It can use a supplied descriptor
// file, protobuf sources, or service reflection to translate JSON or text request data into the
// appropriate protobuf messages and vice versa for presenting the response contents.
package ggrpcurl

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/alsritter/middlebaby/pkg/util/grpcurl"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	// Register gzip compressor so compressed responses will work
	_ "google.golang.org/grpc/encoding/gzip"
	// Register xds so xds and xds-experimental resolver schemes work
	_ "google.golang.org/grpc/xds"
)

// To avoid confusion between program error codes and the gRPC resonse
// status codes 'Cancelled' and 'Unknown', 1 and 2 respectively,
// the response status codes emitted use an offest of 64
const statusCodeOffset = 64

const no_version = "dev build <no version set>"

var version = "grpcurl"

// GGrpCurlDTO 运行GRPCURL DTO
type GGrpCurlDTO struct {
	Plaintext     bool
	FormatError   bool
	EmitDefaults  bool
	AddHeaders    []string
	ImportPaths   []string
	ProtoFiles    []string
	Data          string // 请求数据
	ServiceAddr   string
	ServiceMethod string
	Trace         bool
}

type InvokeGRpc struct {
	connectTimeout     float64
	keepaliveTime      float64
	maxMsgSz           int
	plaintext          bool
	insecure           bool
	cacert             string
	cert               string
	key                string
	serverName         string
	authority          string
	userAgent          string
	data               string
	emitDefaults       bool
	allowUnknownFields bool
	format             string
	addlHeaders        []string
	rpcHeaders         []string
	formatError        bool
	maxTime            float64
	importPaths        []string
	protoFiles         []string

	serviceAddr   string
	serviceMethod string

	trace bool
}

var defaultInvokeGRpc = InvokeGRpc{
	connectTimeout:     0,
	keepaliveTime:      0,
	maxMsgSz:           0,
	plaintext:          false,
	insecure:           false,
	cacert:             "",
	cert:               "",
	key:                "",
	serverName:         "",
	authority:          "",
	userAgent:          "",
	data:               "",
	emitDefaults:       false,
	allowUnknownFields: false,
	format:             "json",
	addlHeaders:        nil,
	rpcHeaders:         nil,
	formatError:        false,
	maxTime:            0,
}

func NewInvokeGRpc(dto *GGrpCurlDTO) *InvokeGRpc {
	clone := defaultInvokeGRpc
	clone.plaintext = dto.Plaintext
	clone.formatError = dto.FormatError
	clone.emitDefaults = dto.EmitDefaults
	clone.addlHeaders = dto.AddHeaders
	clone.importPaths = dto.ImportPaths
	clone.protoFiles = dto.ProtoFiles
	clone.data = dto.Data
	clone.serviceAddr = dto.ServiceAddr
	clone.serviceMethod = dto.ServiceMethod
	clone.trace = dto.Trace
	return &clone
}

// Invoke 运行GGrpCurl 方法 直接发送请求
func (i *InvokeGRpc) Invoke() (metadata.MD, string, *status.Status, error) {

	target := i.serviceAddr
	symbol := i.serviceMethod
	verbosityLevel := 0
	if i.trace {
		verbosityLevel = 2
	}

	ctx := context.Background()
	if i.maxTime > 0 {
		timeout := time.Duration(i.maxTime * float64(time.Second))
		ctx, _ = context.WithTimeout(ctx, timeout)
	}

	dial := i.dial(ctx, target)
	printFormattedStatus := func(w io.Writer, stat *status.Status, formatter grpcurl.Formatter) {
		formattedStatus, err := formatter(stat.Proto())
		if err != nil {
			fmt.Fprintf(w, "ERROR: %v", err.Error())
		}
		fmt.Fprint(w, formattedStatus)
	}

	fileSource, err := grpcurl.DescriptorSourceFromProtoFiles(i.importPaths, i.protoFiles...)
	if err != nil {
		return nil, "", nil, fail(err, "Failed to process proto source files.")
	}

	return i.invoke(dial, verbosityLevel, fileSource, ctx, symbol, printFormattedStatus)
}

func (i *InvokeGRpc) dial(ctx context.Context, target string) func() (*grpc.ClientConn, error) {
	return func() (*grpc.ClientConn, error) {
		dialTime := 10 * time.Second
		if i.connectTimeout > 0 {
			dialTime = time.Duration(i.connectTimeout * float64(time.Second))
		}
		ctx, cancel := context.WithTimeout(ctx, dialTime)
		defer cancel()
		var opts []grpc.DialOption
		if i.keepaliveTime > 0 {
			timeout := time.Duration(i.keepaliveTime * float64(time.Second))
			opts = append(opts, grpc.WithKeepaliveParams(keepalive.ClientParameters{
				Time:    timeout,
				Timeout: timeout,
			}))
		}
		if i.maxMsgSz > 0 {
			opts = append(opts, grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(i.maxMsgSz)))
		}
		var creds credentials.TransportCredentials
		if !i.plaintext {
			var err error
			creds, err = grpcurl.ClientTransportCredentials(i.insecure, i.cacert, i.cert, i.key)
			if err != nil {
				return nil, fail(err, "Failed to configure transport credentials")
			}

			// can use either -servername or -authority; but not both
			if i.serverName != "" && i.authority != "" {
				if i.serverName == i.authority {
					warn("Both -servername and -authority are present; prefer only -authority.")
				} else {
					return nil, fail(nil, "Cannot specify different values for -servername and -authority.")
				}
			}
			overrideName := i.serverName
			if overrideName == "" {
				overrideName = i.authority
			}

			if overrideName != "" {
				if err := creds.OverrideServerName(overrideName); err != nil {
					return nil, fail(err, "Failed to override server name as %q", overrideName)
				}
			}
		} else if i.authority != "" {
			opts = append(opts, grpc.WithAuthority(i.authority))
		}

		grpcurlUA := "grpcurl/" + version
		if version == no_version {
			grpcurlUA = "grpcurl/dev-build (no version set)"
		}
		if i.userAgent != "" {
			grpcurlUA = i.userAgent + " " + grpcurlUA
		}
		opts = append(opts, grpc.WithUserAgent(grpcurlUA))

		network := "tcp"
		cc, err := grpcurl.BlockingDial(ctx, network, target, creds, opts...)
		if err != nil {
			return nil, fail(err, "Failed to dial target host %q", target)
		}
		return cc, nil
	}
}

func (i *InvokeGRpc) invoke(
	dial func() (*grpc.ClientConn, error),
	verbosityLevel int,
	descSource grpcurl.DescriptorSource,
	ctx context.Context,
	symbol string,
	printFormattedStatus func(w io.Writer, stat *status.Status, formatter grpcurl.Formatter),
) (metadata.MD, string, *status.Status, error) {
	cc, err := dial()
	if err != nil {
		return nil, "", nil, err
	}
	defer cc.Close()
	var in = strings.NewReader(i.data)
	// if not verbose output, then also include record delimiters
	// between each message, so output could potentially be piped
	// to another grpcurl process
	includeSeparators := verbosityLevel == 0
	options := grpcurl.FormatOptions{
		EmitJSONDefaultFields: i.emitDefaults,
		IncludeTextSeparator:  includeSeparators,
		AllowUnknownFields:    i.allowUnknownFields,
	}
	rf, formatter, err := grpcurl.RequestParserAndFormatter(grpcurl.Format(i.format), descSource, in, options)
	if err != nil {
		return nil, "", nil, fail(err, "Failed to construct request parser and formatter for %q", i.format)
	}

	var outWrite bytes.Buffer

	h := &CustomEventHandler{
		DefaultEventHandler: &grpcurl.DefaultEventHandler{
			Out:            &outWrite,
			Debug:          os.Stdout,
			Formatter:      formatter,
			VerbosityLevel: verbosityLevel,
		},
	}
	err = grpcurl.InvokeRPC(ctx, descSource, cc, symbol, append(i.addlHeaders, i.rpcHeaders...), h, rf.Next)
	if err != nil {
		if errStatus, ok := status.FromError(err); ok && i.formatError {
			h.Status = errStatus
		} else {
			return nil, "", nil, fail(err, "Error invoking method %q", symbol)
		}
	}
	reqSuffix := ""
	respSuffix := ""
	reqCount := rf.NumRequests()
	if reqCount != 1 {
		reqSuffix = "s"
	}
	if h.NumResponses != 1 {
		respSuffix = "s"
	}
	if verbosityLevel > 0 {
		fmt.Printf("Sent %d request%s and received %d response%s\n", reqCount, reqSuffix, h.NumResponses, respSuffix)
	}
	if h.Status.Code() != codes.OK {
		var errWrite bytes.Buffer
		printFormattedStatus(&errWrite, h.Status, formatter)
		return h.ResponseMd, errWrite.String(), h.Status, nil
	}
	return h.ResponseMd, outWrite.String(), h.Status, nil
}

func warn(msg string, args ...interface{}) {
	msg = fmt.Sprintf("Warning: %s\n", msg)
	fmt.Fprintf(os.Stderr, msg, args...)
}

func fail(err error, msg string, args ...interface{}) error {
	if err != nil {
		msg += ": %v"
		args = append(args, err)
	}
	return fmt.Errorf(msg, args...)
}
