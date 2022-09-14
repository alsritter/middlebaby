package messagepush

import (
	"net/http"
	"sync"
	"time"

	"github.com/alsritter/middlebaby/pkg/util/logger"
	"github.com/gorilla/websocket"
)

// reference:
// * https://github.dev/owenliang/go-push
// * https://www.jianshu.com/p/47876da84627

// WsConnection 连接对象
type WsConnection struct {
	logger.Logger

	ConnId            uint64
	WsSocket          *websocket.Conn
	InChan            chan *WsMessage   // 读队列
	OutChan           chan *PushMessage // 写队列
	mutex             sync.Mutex        // 避免重复关闭通道
	IsClosed          bool              // 是否关闭
	CloseChan         chan byte         // 关闭通知
	isClosed          bool
	lastHeartbeatTime time.Time // 最近一次心跳时间
}

var (
	upGrader = websocket.Upgrader{
		//Allow cross domain
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
)

func (wsConn *WsConnection) WsReadLoop() {
	for {
		msgType, data, err := wsConn.WsSocket.ReadMessage()
		if err != nil {
			goto ERR
		}
		req := &WsMessage{
			msgType,
			data,
		}
		wsConn.Info(nil, "websocket receive request: [%v]", req)

		// 放入请求队列
		select {
		case wsConn.InChan <- req:
		case <-wsConn.CloseChan:
			wsConn.Info(nil, "wsReadLoop close websocket")
			goto CLOSED
		}
	}

ERR:
	wsConn.Close()
CLOSED:
}

func (wsConn *WsConnection) WsWriteLoop() {
	for {
		select {
		//取一个应答
		case msg := <-wsConn.OutChan:
			if err := wsConn.WsSocket.WriteMessage(msg.MessageType, []byte(msg.Content)); err != nil {
				goto ERR
			}
		case <-wsConn.CloseChan:
			goto CLOSED
		}
	}

ERR:
	wsConn.Close()
CLOSED:
}

// 检查心跳（不需要太频繁）
func (wsConn *WsConnection) IsAlive() bool {
	var (
		now = time.Now()
	)

	wsConn.mutex.Lock()
	defer wsConn.mutex.Unlock()

	// 连接已关闭 或者 太久没有心跳
	if wsConn.isClosed || now.Sub(wsConn.lastHeartbeatTime) > 50*time.Second {
		return false
	}
	return true
}

// 更新心跳
func (wsConn *WsConnection) KeepAlive() {
	var (
		now = time.Now()
	)
	wsConn.mutex.Lock()
	defer wsConn.mutex.Unlock()
	wsConn.lastHeartbeatTime = now
}

func (wsConn *WsConnection) Close() {
	wsConn.WsSocket.Close()
	wsConn.mutex.Lock()
	defer wsConn.mutex.Unlock()

	if !wsConn.IsClosed {
		wsConn.IsClosed = true
		close(wsConn.CloseChan)
	}
}
