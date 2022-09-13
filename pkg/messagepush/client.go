package messagepush

import (
	"errors"

	"github.com/alsritter/middlebaby/pkg/util/logger"
	"github.com/gorilla/websocket"
)

// DoConnection websocket
func InitWSConnection(log logger.Logger, connId uint64, wsSocket *websocket.Conn) (wsConnection *WsConnection) {
	// // protocol upgrade.
	// wsSocket, err := upGrader.Upgrade(ctx.Writer, ctx.Request, nil)
	// if err != nil {
	// 	log.Error(nil, "upgrade websocket fail")
	// 	return
	// }

	wsConnection = &WsConnection{
		Logger:    log,
		WsSocket:  wsSocket,
		InChan:    make(chan *WsMessage, 1000),
		OutChan:   make(chan *PushMessage, 1000),
		CloseChan: make(chan byte),
		IsClosed:  false,
	}

	go wsConnection.WsWriteLoop()
	go wsConnection.ProcLoop()
	return
}

func (wsConn *WsConnection) SendMessage(message PushMessage) error {
	select {
	case wsConn.OutChan <- &message:
	case <-wsConn.CloseChan:
		return errors.New("websocket closed")
	}
	return nil
}

func (wsConn *WsConnection) ReadMessage() (*WsMessage, error) {
	select {
	case msg := <-wsConn.InChan:
		return msg, nil
	case <-wsConn.CloseChan:
	}
	return nil, errors.New("websocket closed")
}
