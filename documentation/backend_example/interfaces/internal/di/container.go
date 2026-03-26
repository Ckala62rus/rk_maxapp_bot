package di

import (
	"fmt"
	"log/slog"

	"interfaces/internal/app/handler"
	"interfaces/internal/domain/repository"
	"interfaces/internal/domain/service"
	"interfaces/internal/config"
	"interfaces/pkg"
)

type Container struct {
	Config   *config.Config
	Logger   *slog.Logger
	Minio    *pkg.MinioClient
	Repo     *repository.Repository
	Services *service.Service
	Handler  *handler.Handler
}

type Deps struct {
	Config   *config.Config
	Logger   *slog.Logger
	Minio    *pkg.MinioClient
	Repo     *repository.Repository
	Services *service.Service
	Handler  *handler.Handler
}

func Build(deps Deps) (*Container, error) {
	if deps.Config == nil {
		deps.Config = pkg.MainConfig
	}
	if deps.Logger == nil {
		deps.Logger = pkg.MainLogger
	}

	if deps.Minio == nil {
		minioClient, err := pkg.NewMinioClient(pkg.MinioClientConfig{
			Host:     deps.Config.MinioConfig.Host,
			User:     deps.Config.MinioConfig.User,
			Password: deps.Config.MinioConfig.Password,
			SSL:      deps.Config.MinioConfig.SSL,
		})
		if err != nil {
			return nil, fmt.Errorf("minio init failed: %w", err)
		}
		deps.Minio = minioClient
	}

	if deps.Repo == nil {
		deps.Repo = repository.NewRepository()
	}
	if deps.Services == nil {
		deps.Services = service.NewService(deps.Repo)
	}
	if deps.Handler == nil {
		deps.Handler = handler.NewHandler(deps.Services, deps.Logger, deps.Minio)
	}

	return &Container{
		Config:   deps.Config,
		Logger:   deps.Logger,
		Minio:    deps.Minio,
		Repo:     deps.Repo,
		Services: deps.Services,
		Handler:  deps.Handler,
	}, nil
}

func MustBuild(deps Deps) *Container {
	c, err := Build(deps)
	if err != nil {
		panic(err)
	}
	return c
}
