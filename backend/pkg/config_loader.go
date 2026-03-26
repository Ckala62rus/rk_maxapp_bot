// Package pkg provides application-wide singletons (config, logger, etc.).
package pkg

import (
	"fmt"
	"os"
	"path/filepath"

	"maxapp/internal/config"
)

// MainConfig is loaded once during app startup.
var (
	MainConfig *config.Config
)

// init loads configuration from CONFIG_PATH or fallback to config.yml.
func init() {
	fmt.Println("**** LOAD CONFIG ****")
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		// In local dev, fallback to project dir config.yml.
		projectDir, err := os.Getwd()
		if err != nil {
			panic(err)
		}
		configPath = filepath.Join(projectDir, "config.yml")
	}
	MainConfig = config.MustLoad(configPath)
	fmt.Println("**** CONFIG READY ****")
}
