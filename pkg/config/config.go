package config

import (
	"fmt"
	"net"
	"os"

	"github.com/gookit/slog"
	"go.yaml.in/yaml/v3"
)

type Config struct {
	Server  *Server  `yaml:"server"`
	Message *Message `yaml:"message"`
	Log     *Log     `yaml:"log"`
}

// GenerateDefaultConfig Generates default config for our application.
func GenerateDefaultConfig() *Config {
	return &Config{
		Server:  getDefaultServerConfig(),
		Message: getDefaultMessageConfig(),
		Log:     getDefaultLogConfig(),
	}
}

func ParseConfig(raw string) (*Config, error) {
	cfg := GenerateDefaultConfig()
	// read file
	file, err := os.ReadFile(raw)
	if err != nil {
		return nil, err
	}
	err = yaml.Unmarshal(file, cfg)
	return cfg, err
}

func (cfg Config) Verify() ([]string, []string) {
	errorMsg := make([]string, 0)
	warningMsg := make([]string, 0)

	if ip := net.ParseIP(cfg.Server.BindAddr); ip == nil || ip.To4() == nil {
		errorMsg = append(errorMsg, "Invalid bind address. Note that we only support IPv4 address so far!")
	}
	if cfg.Server.BindPort > 65535 || cfg.Server.BindPort < 1 {
		errorMsg = append(errorMsg, "Invalid bind port number.")
	}
	if cfg.Server.Encrypt.Key == "" {
		errorMsg = append(errorMsg, "Key should not be empty.")
	}
	if len(cfg.Server.Encrypt.Key) < 10 {
		errorMsg = append(errorMsg, "Key must be at least 10 characters.")
	}

	if !cfg.Server.Encrypt.EnableTLS {
		warningMsg = append(warningMsg, "Disabling encryption is not suggested!")
	}

	if cfg.Server.Encrypt.TLSCertPath == "" || cfg.Server.Encrypt.TLSKeyPath == "" {
		warningMsg = append(warningMsg, "Missing cert or key for TLS. Will generate one automatically later.")
	}

	_, err := slog.StringToLevel(cfg.Log.Level)
	if err != nil {
		errorMsg = append(errorMsg, "Provided log level is invalid. You should choose from [ PANIC,FATAL,ERROR,WARNING,NOTICE,INFO,DEBUG,TRACE ].")
	}

	if cfg.Log.LogToFile {
		if cfg.Log.LogFileDirectory == "" {
			warningMsg = append(warningMsg, "Log Directory should not be empty while log to file is enabled. An fallback directory will be applied.")
		} else {
			info, err := os.Stat(cfg.Log.LogFileDirectory)
			if err == nil {
				if !info.IsDir() {
					errorMsg = append(errorMsg, "Log file directory is not a directory.")
				}
			} else {
				errorMsg = append(errorMsg, fmt.Sprintf("Failed to check log file directory: %s", err))
			}
		}
	}
	return warningMsg, errorMsg
}
