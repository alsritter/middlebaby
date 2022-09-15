package messagepush

var (
	G_connMgr *ConnMgr
)

// 推送类型
const (
	PUSH_TYPE_ALL = 2 // 推送在线
)

const (
	bucketCount = 512 // 桶越多, 推送的锁粒度越小, 推送并发度越高

	// 待分发队列的长度
	dispatchChannelSize = 100000 // 分发队列缓冲所有待推送的消息, 等待被分发到Bucket

	// 分发协程的数量
	dispatchWorkerCount = 32 // 每个Bucket有多个协程并发的推送消息

	// Bucket工作队列长度
	bucketJobChannelSize = 1000 // 分发协程用于将待推送消息扇出给各个Bucket

	// Bucket发送协程的数量
	bucketJobWorkerCount = 1000 // 每个Bucket的分发任务放在一个独立队列中
)

// 推送任务
type PushJob struct {
	pushType int    // 推送类型
	roomId   string // 房间ID
	// union {
	bizMsg *BizMessage // 未序列化的业务消息
	wsMsg  *WsMessage  //  已序列化的业务消息
	// }
}

// 连接管理器
type ConnMgr struct {
	buckets []*Bucket
	jobChan []chan *PushJob // 每个Bucket对应一个Job Queue

	dispatchChan chan *PushJob // 待分发消息队列
}

func InitConnMgr() (err error) {
	var (
		bucketIdx         int
		jobWorkerIdx      int
		dispatchWorkerIdx int
		connMgr           *ConnMgr
	)

	connMgr = &ConnMgr{
		buckets:      make([]*Bucket, bucketCount),
		jobChan:      make([]chan *PushJob, bucketCount),
		dispatchChan: make(chan *PushJob, dispatchChannelSize),
	}

	for bucketIdx = range connMgr.buckets {
		connMgr.buckets[bucketIdx] = initBucket(bucketIdx)                     // 初始化Bucket
		connMgr.jobChan[bucketIdx] = make(chan *PushJob, bucketJobChannelSize) // Bucket的Job队列
		// Bucket的Job worker
		for jobWorkerIdx = 0; jobWorkerIdx < bucketJobWorkerCount; jobWorkerIdx++ {
			go connMgr.jobWorkerMain(jobWorkerIdx, bucketIdx)
		}
	}
	// 初始化分发协程, 用于将消息扇出给各个Bucket
	for dispatchWorkerIdx = 0; dispatchWorkerIdx < dispatchWorkerCount; dispatchWorkerIdx++ {
		go connMgr.dispatchWorkerMain(dispatchWorkerIdx)
	}

	G_connMgr = connMgr
	return
}

// 消息分发到Bucket
func (connMgr *ConnMgr) dispatchWorkerMain(dispatchWorkerIdx int) {
	var (
		bucketIdx int
		pushJob   *PushJob
		err       error
	)
	for {
		select {
		case pushJob = <-connMgr.dispatchChan:
			// 序列化
			if pushJob.wsMsg, err = EncodeWSMessage(pushJob.bizMsg); err != nil {
				continue
			}
			// 分发给所有Bucket, 若Bucket拥塞则等待
			for bucketIdx = range connMgr.buckets {
				connMgr.jobChan[bucketIdx] <- pushJob
			}
		}
	}
}

// Job负责消息广播给客户端
func (connMgr *ConnMgr) jobWorkerMain(jobWorkerIdx int, bucketIdx int) {
	var (
		bucket  = connMgr.buckets[bucketIdx]
		pushJob *PushJob
	)

	for {
		select {
		case pushJob = <-connMgr.jobChan[bucketIdx]: // 从Bucket的job queue取出一个任务
			if pushJob.pushType == PUSH_TYPE_ALL {
				bucket.PushAll(pushJob.wsMsg)
			}
		}
	}
}

func (connMgr *ConnMgr) GetBucket(wsConnection *WsConnection) (bucket *Bucket) {
	bucket = connMgr.buckets[wsConnection.ConnId%uint64(len(connMgr.buckets))]
	return
}

func (connMgr *ConnMgr) AddConn(wsConnection *WsConnection) {
	var (
		bucket *Bucket
	)

	bucket = connMgr.GetBucket(wsConnection)
	bucket.AddConn(wsConnection)
}

func (connMgr *ConnMgr) DelConn(wsConnection *WsConnection) {
	var (
		bucket *Bucket
	)

	bucket = connMgr.GetBucket(wsConnection)
	bucket.DelConn(wsConnection)
}

func (connMgr *ConnMgr) PushAll(bizMsg *BizMessage) (err error) {
	var (
		pushJob *PushJob
	)

	pushJob = &PushJob{
		pushType: PUSH_TYPE_ALL,
		bizMsg:   bizMsg,
	}

	select {
	case connMgr.dispatchChan <- pushJob:
	default:
		err = ERR_DISPATCH_CHANNEL_FULL
	}
	return
}
