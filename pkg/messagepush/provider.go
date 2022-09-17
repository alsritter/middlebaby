package messagepush

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/alsritter/middlebaby/pkg/types/msgpush"
	"github.com/alsritter/middlebaby/pkg/util"
	"github.com/alsritter/middlebaby/pkg/util/logger"
	"github.com/alsritter/middlebaby/pkg/util/mbcontext"
	"github.com/gorilla/websocket"
	"github.com/hashicorp/go-multierror"
	"github.com/spf13/pflag"
)

var (
	wsUpgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
)

type Config struct {
	MsgPort        uint32 `json:"msgPort" yaml:"msgPort"`
	WsReadTimeout  uint32 `json:"wsReadTimeout" yaml:"wsReadTimeout"`
	WsWriteTimeout uint32 `json:"wsWriteTimeout" yaml:"wsWriteTimeout"`
}

func NewConfig() *Config {
	return &Config{
		MsgPort:        52162,
		WsReadTimeout:  2000,
		WsWriteTimeout: 2000,
	}
}

func (c *Config) Validate() error {
	if c.MsgPort == 0 {
		return errors.New("[message-push-server] message-push server listener port cannot be empty")
	}

	if c.WsReadTimeout == 0 {
		return errors.New("[message-push-server] message-push server read timeout cannot be empty")
	}

	if c.WsWriteTimeout == 0 {
		return errors.New("[message-push-server] message-push server read timeout cannot be empty")
	}
	return nil
}

// RegisterFlagsWithPrefix is used to register flags
func (c *Config) RegisterFlagsWithPrefix(prefix string, f *pflag.FlagSet) {}

type Provider interface {
	SendMessage(msgpush.PushMessage) error
	Start(ctx *mbcontext.Context) error
}

type msgProvider struct {
	logger.Logger
	cfg *Config

	server     *http.Server
	curConnId  uint64
	lock       sync.Mutex
	connectMap map[uint64]*WsConnection
}

func New(log logger.Logger, cfg *Config) Provider {
	return &msgProvider{
		cfg: cfg,
		server: &http.Server{
			ReadTimeout:  time.Duration(cfg.WsReadTimeout) * time.Millisecond,
			WriteTimeout: time.Duration(cfg.WsWriteTimeout) * time.Millisecond,
		},
		Logger:     log.NewLogger("message-push"),
		connectMap: make(map[uint64]*WsConnection),
	}
}

func (m *msgProvider) SendMessage(message msgpush.PushMessage) error {
	var result error
	m.lock.Lock()
	defer m.lock.Unlock()
	for _, connection := range m.connectMap {
		if connection.IsClosed {
			delete(m.connectMap, connection.ConnId)
			continue
		}

		if err := connection.WsWrite(message); err != nil {
			m.Error(nil, "websocket connId [%d] send message failed, error: [%v]", connection.ConnId, err)
			result = multierror.Append(result, err)
		}
	}
	return result
}

func (m *msgProvider) Start(ctx *mbcontext.Context) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/connect", m.handleConnect)
	m.server.Handler = mux

	util.StartServiceAsync(ctx, m, func() error {
		listener, err := net.Listen("tcp", fmt.Sprintf(":%d", m.cfg.MsgPort))
		if err != nil {
			return fmt.Errorf("failed to listen the port: %d, err: %v", m.cfg.MsgPort, err)
		}

		m.Info(nil, "message-push server started, Listen port: %d", m.cfg.MsgPort)
		if err := m.server.Serve(listener); err != nil {
			if err.Error() != "http: Server closed" {
				return fmt.Errorf("failed to start the message-push server: %v", err)
			}
		}

		return nil
	}, func() error {
		m.Info(nil, "stopping server...")
		if err := m.server.Shutdown(context.TODO()); err != nil {
			return fmt.Errorf("server Shutdown failed: [%v]", err)
		}
		return nil
	})

	return nil
}

func (m *msgProvider) handleConnect(resp http.ResponseWriter, req *http.Request) {
	var (
		err      error
		wsSocket *websocket.Conn
		connId   uint64
		wsConn   *WsConnection
	)

	if wsSocket, err = wsUpgrader.Upgrade(resp, req, nil); err != nil {
		m.Error(nil, "websocket connection field, error: [%v]", err)
		return
	}

	connId = atomic.AddUint64(&m.curConnId, 1)
	wsConn = initWSConnection(logger.NewDefault("websocket"), connId, wsSocket)
	m.lock.Lock()
	m.connectMap[connId] = wsConn
	m.lock.Unlock()
}
