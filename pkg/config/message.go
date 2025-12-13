package config

type MsgType string

const (
	DecodedMsg MsgType = "DECODED"
	LoggedMsg  MsgType = "LOGGED"
	RigMsg     MsgType = "RIG"
)

type Message struct {
	ForwardMessageType []MsgType `yaml:"forward_message_type"`
}

func getDefaultMessageConfig() *Message {
	return &Message{
		ForwardMessageType: []MsgType{DecodedMsg, LoggedMsg, RigMsg},
	}
}
