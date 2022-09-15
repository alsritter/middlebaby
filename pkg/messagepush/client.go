package messagepush

import "encoding/json"

func InitMessagePush() error {
	if err := initWSServer(); err != nil {
		return err
	}

	if err := initConnMgr(); err != nil {
		return err
	}

	return nil
}

func SendMessage(val interface{}) error {
	data, err := json.Marshal(val)
	if err != nil {
		return err
	}

	if err := g_connMgr.PushAll(&BizMessage{
		Type: "PUSH",
		Data: data,
	}); err != nil {
		return err
	}

	return nil
}
