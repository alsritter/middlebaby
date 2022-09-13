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

type WsMessage struct {
	messageType int
	data        []byte
}

// PushMessage 推送消息内容
type PushMessage struct {
	ID          int    `json:"id" yaml:"id"`                   // 编号
	Extra       string `json:"extra" yaml:"extra"`             // 额外信息
	MessageType int    `json:"messageType" yaml:"messageType"` // 消息类型
	Content     string `json:"content" yaml:"content"`         // 消息内容
}

// WsConnection 连接对象
type WsConnection struct {
	logger.Logger

	ConnId    uint64
	WsSocket  *websocket.Conn
	InChan    chan *WsMessage   // 读队列
	OutChan   chan *PushMessage // 写队列
	Mutex     sync.Mutex        // 避免重复关闭通道
	IsClosed  bool              // 是否关闭
	CloseChan chan byte         // 关闭通知
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
	wsConn.wsClose()
CLOSED:
}

func (wsConn *WsConnection) WsWriteLoop() {
	for {
		select {
		//取一个应答
		case msg := <-wsConn.OutChan:
			if err := wsConn.WsSocket.WriteJSON(msg); err != nil {
				goto ERR
			}
		case <-wsConn.CloseChan:
			goto CLOSED
		}
	}

ERR:
	wsConn.wsClose()
CLOSED:
}

// ProcLoop 心跳检测
func (wsConn *WsConnection) ProcLoop() {
	// 启动一个goroutine 发送心跳
	go func() {
		for {
			time.Sleep(50 * time.Second)
			if err := wsConn.WsSocket.WriteMessage(websocket.PingMessage, []byte("heartbeat")); err != nil {
				wsConn.Info(nil, "heartbeat fail [%v]", err)
				wsConn.wsClose()
				break
			}
		}
	}()
}

func (wsConn *WsConnection) wsClose() {
	wsConn.WsSocket.Close()
	wsConn.Mutex.Lock()
	defer wsConn.Mutex.Unlock()
	if !wsConn.IsClosed {
		wsConn.IsClosed = true
		close(wsConn.CloseChan)
	}
}
