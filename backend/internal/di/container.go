// Package di wires dependencies into a container.
package di

import (
	"log/slog"
	"strings"

	"maxapp/internal/app/handler"
	"maxapp/internal/config"
	"maxapp/internal/domain"
	"maxapp/internal/infrastructure/mssql"
	"maxapp/internal/infrastructure/postgres"
	"maxapp/internal/service"
)

// Container aggregates constructed dependencies.
type Container struct {
	Config  *config.Config
	Logger  *slog.Logger
	Handler *handler.Handler
}

// Build creates repositories, services and handlers.
func Build(cfg *config.Config, logger *slog.Logger) (*Container, error) {
	// Connect to Postgres first (users + history).
	pg, err := postgres.New(cfg.Postgres)
	if err != nil {
		return nil, err
	}
	logger.Info("postgres connected", "host", cfg.Postgres.Host, "db", cfg.Postgres.Database)

	// Connect to MSSQL (DAX).
	var warehouseRepo domain.WarehouseRepository
	ms, err := mssql.New(cfg.MSSQL)
	if err != nil {
		// In development we allow startup without MSSQL (search will return error).
		if strings.EqualFold(cfg.Env, "development") {
			logger.Error("mssql connect failed, continuing in dev", "error", err)
			warehouseRepo = mssql.NewUnavailableWarehouseRepository(err)
		} else {
			return nil, err
		}
	} else {
		logger.Info("mssql connected", "host", cfg.MSSQL.Host, "db", cfg.MSSQL.Database)
		warehouseRepo = mssql.NewWarehouseRepository(ms)
	}

	// Repositories.
	userRepo := postgres.NewUserRepository(pg)
	historyRepo := postgres.NewSearchHistoryRepository(pg)
	// Services.
	authService := service.NewAuthService(userRepo, cfg.Max.BotToken, logger)
	userService := service.NewUserService(userRepo, logger)
	warehouseService := service.NewWarehouseService(warehouseRepo, historyRepo, logger)

	// HTTP handlers.
	h := handler.NewHandler(authService, userService, warehouseService, logger, cfg.Max.AllowMockInitData)

	return &Container{
		Config:  cfg,
		Logger:  logger,
		Handler: h,
	}, nil
}
