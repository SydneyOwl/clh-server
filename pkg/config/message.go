package config

type MsgType string

const (
	DecodedMsg MsgType = "DECODED"
	LoggedMsg  MsgType = "LOGGED"
	RigMsg     MsgType = "RIG"
)

type Message struct {
	Key                string    `yaml:"key"`
	EnableTLS          bool      `yaml:"encrypt"`
	TLSCertPath        string    `yaml:"tls_cert_path"`
	TLSKeyPath         string    `yaml:"tls_key_path"`
	TLSCACertPath      string    `yaml:"tls_ca_cert_path"`
	ForwardMessageType []MsgType `yaml:"forward_message_type"`
}

func getDefaultMessageConfig() *Message {
	return &Message{
		ForwardMessageType: []MsgType{DecodedMsg, LoggedMsg, RigMsg},
		EnableTLS:          true,
		Key:                "",
	}
}
