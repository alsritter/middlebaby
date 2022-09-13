package messagepush

import (
	"net/http"
	"sync"

	"github.com/alsritter/middlebaby/pkg/util/logger"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var (
	upGrader = websocket.Upgrader{
		// 允许跨域
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}

	mutex sync.RWMutex

	// Relation save websocket connection
	Relation = make(map[string]WsConnection)
)

// DoConnection websocket 协议
// Reference: https://www.jianshu.com/p/47876da84627
func DoConnection(log logger.Logger, ctx *gin.Context) {
	connectionNo := ctx.GetString("connectionNo")
	if connectionNo == "" {
		ctx.JSON(http.StatusUnprocessableEntity, http.StatusText(http.StatusUnprocessableEntity))
		return
	}

	// 升级为WebSocket协议
	wsSocket, err := upGrader.Upgrade(ctx.Writer, ctx.Request, nil)
	if err != nil {
		log.Error(nil, "upgrade websocket fail")
		return
	}

	// 添加连接对应关系
	handleConnect(connectionNo, WsConnection{
		Logger:    log,
		WsSocket:  wsSocket,
		InChan:    make(chan *WsMessage, 1000),
		OutChan:   make(chan *PushMessage, 1000),
		CloseChan: make(chan byte),
		IsClosed:  false,
	})
}

func handleConnect(connectionNo string, wsConn WsConnection) {
	mutex.Lock()
	Relation[connectionNo] = wsConn
	mutex.Unlock()
	go wsConn.WsWriteLoop()
	go wsConn.ProcLoop()
}
