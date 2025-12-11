package config

type LogLevel string

type Log struct {
	LogToFile        bool   `yaml:"log_to_file"`
	Level            string `yaml:"level"`
	LogFileDirectory string `yaml:"log_file_directory"`
}

func getDefaultLogConfig() *Log {
	return &Log{
		LogToFile:        false,
		Level:            "Info",
		LogFileDirectory: "",
	}
}
