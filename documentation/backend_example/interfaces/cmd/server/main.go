package main

import (
	"context"
	"log"

	"interfaces/internal/di"
	"interfaces/internal/profiling"
	"interfaces/internal/server"
	"interfaces/internal/telemetry"
	"interfaces/pkg"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

func main() {
	logger := pkg.MainLogger
	logger.Info("*** Start program ***")

	cfg := pkg.MainConfig

	shutdownTelemetry, err := telemetry.Init(context.Background(), telemetry.LoadConfigFromEnv())
	if err != nil {
		logger.Error("failed to init telemetry: " + err.Error())
	} else {
		defer func() {
			_ = shutdownTelemetry(context.Background())
		}()
	}

	profiler, err := profiling.Start(profiling.LoadConfigFromEnv())
	if err != nil {
		logger.Error("failed to init profiling: " + err.Error())
	} else if profiler != nil {
		defer profiler.Stop()
	}

	container, err := di.Build(di.Deps{
		Config: cfg,
		Logger: logger,
	})
	if err != nil {
		log.Fatal(err)
	}

	srv := new(server.Server)
	otelHandler := otelhttp.NewHandler(container.Handler.InitRoutes(), "http-server")
	err = srv.Run(container.Config.HttpServer.Port, otelHandler)
	if err != nil {
		logger.Error(err.Error())
	}

	//user, err := services.GetUserById(2)
	//if err != nil {
	//	log.Fatal(err.Error())
	//}
	//fmt.Println(user.Email)
}
