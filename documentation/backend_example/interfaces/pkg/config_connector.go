package pkg

import (
	"fmt"
	"interfaces/internal/config"
	"os"
	"path/filepath"
)

var (
	MainConfig *config.Config
)

func init() {
	fmt.Println("***************** LOAD CONFIG *****************")
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		projectDir, err := os.Getwd()
		if err != nil {
			panic(err)
		}
		configPath = filepath.Join(projectDir, "config.yml")
	}
	MainConfig = config.MustLoad(configPath)
	fmt.Println("***************** CONFIG RUNNING *****************")
}
