package messagepush

import (
	"encoding/json"

	"github.com/gorilla/websocket"
)

type WsMessage struct {
	messageType int
	data        []byte
}

// 业务消息的固定格式 (type + data)
type BizMessage struct {
	Type string          `json:"type"` // type消息类型: PING, PONG, PUSH
	Data json.RawMessage `json:"data"` // data数据字段
}

// Data数据类型

// PING
type BizPingData struct{}

// PONG
type BizPongData struct{}

func EncodeWSMessage(bizMessage *BizMessage) (wsMessage *WsMessage, err error) {
	var (
		buf []byte
	)
	if buf, err = json.Marshal(*bizMessage); err != nil {
		return
	}
	wsMessage = &WsMessage{websocket.TextMessage, buf}
	return
}

// 解析{"type": "PING", "data": {...}}的包
func DecodeBizMessage(buf []byte) (bizMessage *BizMessage, err error) {
	var (
		bizMsgObj BizMessage
	)

	if err = json.Unmarshal(buf, &bizMsgObj); err != nil {
		return
	}

	bizMessage = &bizMsgObj
	return
}
