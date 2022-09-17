package messagepush

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/alsritter/middlebaby/pkg/types/msgpush"
	"github.com/alsritter/middlebaby/pkg/util/logger"
	"github.com/gorilla/websocket"
)

// WsConnection 连接对象
type WsConnection struct {
	logger.Logger

	ConnId    uint64
	WsSocket  *websocket.Conn
	InChan    chan *msgpush.WsMessage   // 读队列
	OutChan   chan *msgpush.PushMessage // 写队列
	Mutex     sync.Mutex                // 避免重复关闭通道
	IsClosed  bool                      // 是否关闭
	CloseChan chan byte                 // 关闭通知
}

// DoConnection websocket
func initWSConnection(log logger.Logger, connId uint64, wsSocket *websocket.Conn) (wsConnection *WsConnection) {
	wsConnection = &WsConnection{
		Logger:    log,
		ConnId:    connId,
		WsSocket:  wsSocket,
		InChan:    make(chan *msgpush.WsMessage, 1000),
		OutChan:   make(chan *msgpush.PushMessage, 1000),
		CloseChan: make(chan byte),
		IsClosed:  false,
	}

	go wsConnection.WsWriteLoop()
	go wsConnection.WsReadLoop()
	go wsConnection.ProcLoop()
	return
}

// WsReadLoop ...
func (wsConn *WsConnection) WsReadLoop() {
	for {
		msgType, data, err := wsConn.WsSocket.ReadMessage()
		if err != nil {
			goto ERROR
		}
		req := &msgpush.WsMessage{
			MessageType: msgType,
			Data:        data,
		}
		fmt.Println(req)
		// 放入请求队列
		select {
		case wsConn.InChan <- req:
		case <-wsConn.CloseChan:
			wsConn.Info(nil, "wsReadLoop close websocket")
			goto CLOSED
		}
	}
ERROR:
	wsConn.wsClose()
CLOSED:
}

// WsWriteLoop ...
func (wsConn *WsConnection) WsWriteLoop() {
	for {
		select {
		//取一个应答
		case msg := <-wsConn.OutChan:
			if err := wsConn.WsSocket.WriteJSON(msg); err != nil {
				goto ERROR
			}
		case <-wsConn.CloseChan:
			goto CLOSED
		}
	}

ERROR:
	wsConn.wsClose()
CLOSED:
}

// WsWrite ...
func (wsConn *WsConnection) WsWrite(message msgpush.PushMessage) error {
	select {
	case wsConn.OutChan <- &message:
	case <-wsConn.CloseChan:
		return errors.New("websocket closed")
	}
	return nil
}

// WsRead ...
func (wsConn *WsConnection) WsRead() (*msgpush.WsMessage, error) {
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
				wsConn.Error(nil, "heartbeat fail %v", err.Error())
				wsConn.wsClose()
				break
			}
		}
	}()
}

func (wsConn *WsConnection) wsClose() {
	if err := wsConn.WsSocket.Close(); err != nil {
		wsConn.Error(nil, "websocket close fail [%v]", err)
	}
	wsConn.Mutex.Lock()
	defer wsConn.Mutex.Unlock()
	if !wsConn.IsClosed {
		wsConn.IsClosed = true
		close(wsConn.CloseChan)
	}
}
