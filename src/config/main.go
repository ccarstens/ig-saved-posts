package config

import (
	"os"
	"path"
)

const (
	configFile = ".ig-saved-posts/config.json"
)

func GetConfigFilePath() string {
	home, _ := os.UserHomeDir()
	return path.Join(home, configFile)
}
