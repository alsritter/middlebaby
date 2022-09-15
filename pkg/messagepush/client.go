package messagepush

import (
	"time"

	"github.com/alsritter/middlebaby/pkg/util/logger"
	"github.com/gorilla/websocket"
)

// DoConnection websocket
func initWSConnection(log logger.Logger, connId uint64, wsSocket *websocket.Conn) (wsConnection *WsConnection) {
	wsConnection = &WsConnection{
		Logger:            log,
		WsSocket:          wsSocket,
		InChan:            make(chan *WsMessage, 1000),
		OutChan:           make(chan *WsMessage, 1000),
		CloseChan:         make(chan byte),
		lastHeartbeatTime: time.Now(),
		IsClosed:          false,
	}

	go wsConnection.WsWriteLoop()
	go wsConnection.WsReadLoop()
	return
}

func (wsConn *WsConnection) SendMessage(message *WsMessage) error {
	select {
	case wsConn.OutChan <- message:
	case <-wsConn.CloseChan:
		return ERR_CONNECTION_CLOSED
	default:
		return ERR_SEND_MESSAGE_FULL
	}
	return nil
}

func (wsConn *WsConnection) ReadMessage() (*WsMessage, error) {
	select {
	case msg := <-wsConn.InChan:
		return msg, nil
	case <-wsConn.CloseChan:
	}
	return nil, ERR_CONNECTION_CLOSED
}
