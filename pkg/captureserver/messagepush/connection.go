package messagepush

import (
	"errors"
	"sync"
	"time"

	"github.com/alsritter/middlebaby/pkg/util/logger"
	"github.com/gorilla/websocket"
)

type WsMessage struct {
	messageType int
	data        []byte
}

// PushMessage 推送消息内容
type PushMessage struct {
	ID          int    `json:"id"`           // 编号
	ReadFlag    int    `json:"read_flag"`    // 已读标记
	Extra       string `json:"extra"`        // 额外信息
	MessageType int    `json:"message_type"` // 消息类型
	Content     string `json:"content"`      // 消息内容
	Count       int    `json:"count"`        // 总数
}

// WsConnection 连接对象
type WsConnection struct {
	logger.Logger

	WsSocket  *websocket.Conn
	InChan    chan *WsMessage   // 读队列
	OutChan   chan *PushMessage // 写队列
	Mutex     sync.Mutex        // 避免重复关闭通道
	IsClosed  bool              // 是否关闭
	CloseChan chan byte         // 关闭通知
}

func (wsConn *WsConnection) WsReadLoop() {
	for {
		msgType, data, err := wsConn.WsSocket.ReadMessage()
		if err != nil {
			goto toError
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
			goto toClosed
		}
	}

toError:
	wsConn.wsClose()
toClosed:
}

func (wsConn *WsConnection) WsWriteLoop() {
	for {
		select {
		//取一个应答
		case msg := <-wsConn.OutChan:
			if err := wsConn.WsSocket.WriteJSON(msg); err != nil {
				goto toError
			}
		case <-wsConn.CloseChan:
			goto toClosed
		}
	}

toError:
	wsConn.wsClose()
toClosed:
}

func (wsConn *WsConnection) WsWrite(message PushMessage) error {
	select {
	case wsConn.OutChan <- &message:
	case <-wsConn.CloseChan:
		return errors.New("websocket closed")
	}
	return nil
}

func (wsConn *WsConnection) WsRead() (*WsMessage, error) {
	select {
	case msg := <-wsConn.InChan:
		return msg, nil
	case <-wsConn.CloseChan:
	}
	return nil, errors.New("websocket closed")
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
