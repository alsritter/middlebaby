package task

type RunTaskReply struct {
	Status       int32  `yaml:"status" json:"status"`
	FailedReason string `yaml:"failedReason" json:"failedReason"`
}
