package messagepush

import "encoding/json"

type WsMessage struct {
	messageType int
	data        []byte
}

// 业务消息的固定格式(type+data)
type BizMessage struct {
	Type string          `json:"type"` // type消息类型: PING, PONG, JOIN, LEAVE, PUSH
	Data json.RawMessage `json:"data"` // data数据字段
}

// PushMessage 推送消息内容
type PushMessage struct {
	ID          int    `json:"id" yaml:"id"`                   // 编号
	Extra       string `json:"extra" yaml:"extra"`             // 额外信息
	MessageType int    `json:"messageType" yaml:"messageType"` // 消息类型
	Content     string `json:"content" yaml:"content"`         // 消息内容
}
