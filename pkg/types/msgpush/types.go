package msgpush

type MsgType string

const (
	Capture MsgType = "capture"
)

// WsMessage ...
type WsMessage struct {
	MessageType int
	Data        []byte
}

// PushMessage 推送消息内容
type PushMessage struct {
	ID          int     `json:"id"`          // 编号
	Extra       string  `json:"extra"`       // 额外信息
	MessageType MsgType `json:"messageType"` // 消息类型
	Content     string  `json:"content"`     // 消息内容
}
