package config

import (
	"fmt"
	"os"
	"path/filepath"
)

type PathConfig struct {
	Root        string
	LogDir      string
	SessionsDir string
}

// GetDefaultPathConfig initializes and returns the default directory configuration for the application.
func GetDefaultPathConfig() (*PathConfig, error) {

	baseDir, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}
	root := filepath.Join(baseDir, ".tinycoder")

	paths := &PathConfig{
		Root:        root,
		LogDir:      filepath.Join(root, "logs"),
		SessionsDir: filepath.Join(root, "sessions"),
	}

	// 3. Ensure all directories exist on disk
	dirsToCreate := []string{
		paths.LogDir,
		paths.SessionsDir,
	}

	for _, dir := range dirsToCreate {
		// MkdirAll does nothing and returns nil if the directory already exists
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	return paths, nil
}
