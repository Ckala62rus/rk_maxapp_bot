// Package profiling wires Pyroscope continuous profiling.
package profiling

import (
	"os"
	"strings"

	pyroscope "github.com/grafana/pyroscope-go"
)

// Config describes profiling settings loaded from env.
type Config struct {
	Enabled bool
	Server  string
	AppName string
	Env     string
}

// LoadConfigFromEnv reads profiling settings from environment.
func LoadConfigFromEnv() Config {
	enabled := strings.EqualFold(os.Getenv("ENABLE_PROFILING"), "true")
	server := os.Getenv("PYROSCOPE_URL")
	if server == "" {
		server = "http://pyroscope:4040"
	}
	appName := os.Getenv("APP_NAME")
	if appName == "" {
		appName = "maxapp"
	}
	env := os.Getenv("ENV")
	if env == "" {
		env = "development"
	}

	return Config{
		Enabled: enabled,
		Server:  server,
		AppName: appName,
		Env:     env,
	}
}

// Start initializes Pyroscope profiler or returns nil when disabled.
func Start(cfg Config) (*pyroscope.Profiler, error) {
	if !cfg.Enabled {
		// Profiling disabled, return nil without error.
		return nil, nil
	}

	// Configure multiple profile types for deeper analysis.
	return pyroscope.Start(pyroscope.Config{
		ApplicationName: cfg.AppName,
		ServerAddress:   cfg.Server,
		Tags: map[string]string{
			"env": cfg.Env,
		},
		ProfileTypes: []pyroscope.ProfileType{
			pyroscope.ProfileCPU,
			pyroscope.ProfileInuseObjects,
			pyroscope.ProfileAllocObjects,
			pyroscope.ProfileInuseSpace,
			pyroscope.ProfileAllocSpace,
			pyroscope.ProfileGoroutines,
			pyroscope.ProfileMutexCount,
			pyroscope.ProfileMutexDuration,
			pyroscope.ProfileBlockCount,
			pyroscope.ProfileBlockDuration,
		},
	})
}
