package messagepush

import "net/http"

type WsServer struct {
	server    *http.Server
	curConnId uint64
}

func (s *WsServer) Start() {

}
