package config

type LogConfig struct {
	LogFileName string
	LogLevel    string
	MaxSizeMB   int
	MaxBackups  int
	MaxAgeDays  int
	Compress    bool
}

func GetDefaultLogConfig() *LogConfig {
	return &LogConfig{
		LogFileName: "tinycoder.jsonl",
		LogLevel:    "info",
		MaxSizeMB:   5,
		MaxBackups:  5,
		MaxAgeDays:  14,
		Compress:    false,
	}
}
