package messagepush

import (
	"encoding/json"
	"time"

	"github.com/gorilla/websocket"
)

// 处理websocket请求
func (wsConn *WsConnection) WSHandle() {
	var (
		message *WsMessage
		bizReq  *BizMessage
		bizResp *BizMessage
		err     error
		buf     []byte
	)

	// 连接加入管理器, 可以推送端查找到
	G_connMgr.AddConn(wsConn)

	// 心跳检测线程
	go wsConn.heartbeatChecker()

	// 请求处理协程
	for {
		if message, err = wsConn.ReadMessage(); err != nil {
			goto ERR
		}

		// 只处理文本消息
		if message.messageType != websocket.TextMessage {
			continue
		}

		// 解析消息体
		if bizReq, err = DecodeBizMessage(message.data); err != nil {
			goto ERR
		}

		bizResp = nil

		// 请求串行处理
		switch bizReq.Type {
		// 收到PING则响应PONG: {"type": "PING"}, {"type": "PONG"}
		case "PING":
			if bizResp, err = wsConn.handlePing(bizReq); err != nil {
				goto ERR
			}
		}

		if bizResp != nil {
			if buf, err = json.Marshal(*bizResp); err != nil {
				goto ERR
			}
			// socket缓冲区写满不是致命错误
			if err = wsConn.SendMessage(&WsMessage{websocket.TextMessage, buf}); err != nil {
				if err != ERR_SEND_MESSAGE_FULL {
					goto ERR
				} else {
					err = nil
				}
			}
		}
	}

ERR:
	// 确保连接关闭
	wsConn.Close()
	// 从连接池中移除
	G_connMgr.DelConn(wsConn)
	return
}

// 每隔1秒, 检查一次连接是否健康
func (wsConn *WsConnection) heartbeatChecker() {
	var (
		timer *time.Timer
	)
	timer = time.NewTimer(50 * time.Second)
	for {
		select {
		case <-timer.C:
			if !wsConn.IsAlive() {
				wsConn.Close()
				goto EXIT
			}
			timer.Reset(50 * time.Second)
		case <-wsConn.CloseChan:
			timer.Stop()
			goto EXIT
		}
	}

EXIT:
	// 确保连接被关闭
}

// 处理PING请求
func (wsConn *WsConnection) handlePing(bizReq *BizMessage) (bizResp *BizMessage, err error) {
	var (
		buf []byte
	)

	wsConn.KeepAlive()

	if buf, err = json.Marshal(BizPongData{}); err != nil {
		return
	}
	bizResp = &BizMessage{
		Type: "PONG",
		Data: json.RawMessage(buf),
	}
	return
}
