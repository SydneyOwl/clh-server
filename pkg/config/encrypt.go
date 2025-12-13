package config

type Encrypt struct {
	Key           string `yaml:"key"`
	EnableTLS     bool   `yaml:"encrypt"`
	TLSCertPath   string `yaml:"tls_cert_path"`
	TLSKeyPath    string `yaml:"tls_key_path"`
	TLSCACertPath string `yaml:"tls_ca_cert_path"`
}

func getDefaultEncryptConfig() *Encrypt {
	return &Encrypt{
		Key:           "",
		EnableTLS:     true,
		TLSCertPath:   "",
		TLSKeyPath:    "",
		TLSCACertPath: "",
	}
}
