package config

type Server struct {
	BindAddr  string     `yaml:"bind_addr"`
	BindPort  int        `yaml:"bind_port"`
	Encrypt   *Encrypt   `yaml:"encrypt"`
	Transport *Transport `yaml:"transport"`
}

func getDefaultServerConfig() *Server {
	return &Server{
		BindAddr:  "0.0.0.0",
		BindPort:  7410,
		Encrypt:   getDefaultEncryptConfig(),
		Transport: getDefaultTransportConfig(),
	}
}
