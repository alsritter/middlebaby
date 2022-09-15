package messagepush

import (
	"net"
	"net/http"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/alsritter/middlebaby/pkg/util/logger"
	"github.com/gorilla/websocket"
)

const (
	wsReadTimeout  = 2000
	wsWriteTimeout = 2000
)

var (
	g_wsServer *WsServer

	wsUpgrader = websocket.Upgrader{
		// 允许所有CORS跨域请求
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
)

type WsServer struct {
	server    *http.Server
	curConnId uint64
}

func initWSServer() (err error) {
	var (
		mux      *http.ServeMux
		server   *http.Server
		listener net.Listener
	)

	// 路由
	mux = http.NewServeMux()
	mux.HandleFunc("/connect", handleConnect)

	// HTTP服务
	server = &http.Server{
		ReadTimeout:  time.Duration(wsReadTimeout) * time.Millisecond,
		WriteTimeout: time.Duration(wsWriteTimeout) * time.Millisecond,
		Handler:      mux,
	}

	// 监听端口
	if listener, err = net.Listen("tcp", ":"+strconv.Itoa(56271)); err != nil {
		return
	}

	// 赋值全局变量
	g_wsServer = &WsServer{
		server:    server,
		curConnId: uint64(time.Now().Unix()),
	}

	// 拉起服务
	go server.Serve(listener)
	return
}

func handleConnect(resp http.ResponseWriter, req *http.Request) {
	var (
		err      error
		wsSocket *websocket.Conn
		connId   uint64
		wsConn   *WsConnection
	)

	// WebSocket握手
	if wsSocket, err = wsUpgrader.Upgrade(resp, req, nil); err != nil {
		return
	}

	// 连接唯一标识
	connId = atomic.AddUint64(&g_wsServer.curConnId, 1)

	// 初始化WebSocket的读写协程
	wsConn = initWSConnection(logger.NewDefault("websocket"), connId, wsSocket)

	// 开始处理websocket消息
	wsConn.WSHandle()
}
