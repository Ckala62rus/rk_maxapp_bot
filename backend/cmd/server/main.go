// Package main contains application entrypoint.
package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Ckala62rus/rk_maxapp_bot/backend/internal/di"
	"github.com/Ckala62rus/rk_maxapp_bot/backend/internal/profiling"
	"github.com/Ckala62rus/rk_maxapp_bot/backend/internal/telemetry"
	"github.com/Ckala62rus/rk_maxapp_bot/backend/pkg"
	"github.com/Ckala62rus/rk_maxapp_bot/backend/pkg/logger"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

// main bootstraps config, observability, DI container and HTTP server.
func main() {
	// Load config first, it is required for all subsequent components.
	cfg := pkg.MainConfig

	// Initialize logger with level/env from config.
	logg := logger.New(logger.Config{Level: cfg.Log.Level, Env: cfg.Env})
	logg.Info("service starting", "env", cfg.Env, "log_level", cfg.Log.Level)

	// Initialize tracing (OpenTelemetry) and log the result.
	telemetryCfg := telemetry.LoadConfigFromEnv()
	shutdownTelemetry, err := telemetry.Init(context.Background(), telemetryCfg)
	if err != nil {
		logg.Error("telemetry init failed", "error", err)
	} else {
	logg.Info("telemetry initialized", "enabled", telemetryCfg.Enabled, "endpoint", telemetryCfg.Endpoint)
		defer func() {
			_ = shutdownTelemetry(context.Background())
		}()
	}

	// Initialize profiling (Pyroscope) and log the result.
	profilingCfg := profiling.LoadConfigFromEnv()
	profiler, err := profiling.Start(profilingCfg)
	if err != nil {
		logg.Error("profiling init failed", "error", err)
	} else {
	logg.Info("profiling initialized", "enabled", profilingCfg.Enabled, "server", profilingCfg.Server)
	}
	if profiler != nil {
		defer profiler.Stop()
	}

	// Build DI container (DB connections, repositories, services, handlers).
	container, err := di.Build(cfg, logg)
	if err != nil {
		log.Fatal(err)
	}
	logg.Info("container initialized")
	logg.Info("metrics endpoint ready", "path", "/metrics")

	// Compose address from host/port.
	addr := cfg.HttpServer.Address
	if addr == "" {
		addr = "0.0.0.0"
	}
	if cfg.HttpServer.Port != "" {
		addr = fmt.Sprintf("%s:%s", addr, cfg.HttpServer.Port)
	}

	// Wrap router with otelhttp to get request spans.
	handler := otelhttp.NewHandler(container.Handler.InitRoutes(), "http-server")
	server := &http.Server{
		Addr:         addr,
		Handler:      handler,
		ReadTimeout:  cfg.HttpServer.Timeout,
		WriteTimeout: cfg.HttpServer.Timeout,
		IdleTimeout:  cfg.HttpServer.IdleTimeout,
	}

	// Start HTTP server in background.
	go func() {
		logg.Info("http server started", "addr", addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logg.Error("http server error", "error", err)
		}
	}()

	// Graceful shutdown on SIGINT/SIGTERM.
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logg.Error("server shutdown error", "error", err)
	}

	logg.Info("service stopped")
}
