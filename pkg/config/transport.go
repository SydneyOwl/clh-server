package config

type Transport struct {
	HeartbeatTimeoutSec float64 `yaml:"heartbeat_timeout_sec"`
}

func getDefaultTransportConfig() *Transport {
	return &Transport{
		HeartbeatTimeoutSec: 16,
	}
}
