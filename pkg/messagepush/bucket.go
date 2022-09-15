package messagepush

import "sync"

type Bucket struct {
	rwMutex sync.RWMutex
	index   int                      // 我是第几个桶
	id2Conn map[uint64]*WsConnection // 连接列表(key=连接唯一ID)
}

func initBucket(bucketIdx int) (bucket *Bucket) {
	bucket = &Bucket{
		index:   bucketIdx,
		id2Conn: make(map[uint64]*WsConnection),
	}
	return
}

func (bucket *Bucket) AddConn(wsConn *WsConnection) {
	bucket.rwMutex.Lock()
	defer bucket.rwMutex.Unlock()

	bucket.id2Conn[wsConn.ConnId] = wsConn
}

func (bucket *Bucket) DelConn(wsConn *WsConnection) {
	bucket.rwMutex.Lock()
	defer bucket.rwMutex.Unlock()

	delete(bucket.id2Conn, wsConn.ConnId)
}

// 推送给Bucket内所有用户
func (bucket *Bucket) PushAll(wsMsg *WsMessage) {
	var (
		wsConn *WsConnection
	)

	// 锁Bucket
	bucket.rwMutex.RLock()
	defer bucket.rwMutex.RUnlock()

	// 全量非阻塞推送
	for _, wsConn = range bucket.id2Conn {
		wsConn.SendMessage(wsMsg)
	}
}
