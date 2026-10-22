package config

type AppConfig struct {
	PathConfig *PathConfig
	LogConfig  *LogConfig
}

// GetDefaultAppConfig initializes and returns the default application configuration, including path and log settings.
func GetDefaultAppConfig() (*AppConfig, error) {
	path, err := GetDefaultPathConfig()
	if err != nil {
		return nil, err
	}
	return &AppConfig{
		PathConfig: path,
		LogConfig:  GetDefaultLogConfig(),
	}, nil
}
