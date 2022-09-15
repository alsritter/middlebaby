package messagepush

func InitMessagePush() error {
	if err := initWSServer(); err != nil {
		return err
	}

	if err := initConnMgr(); err != nil {
		return err
	}

	return nil
}

func SendMessage(val string) error {
	if err := g_connMgr.PushAll(&BizMessage{
		Type: "PUSH",
		Data: []byte(val),
	}); err != nil {
		return err
	}

	return nil
}
