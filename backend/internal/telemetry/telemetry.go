// Package telemetry wires OpenTelemetry tracing.
package telemetry

import (
	"context"
	"os"
	"strings"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.25.0"
)

// Config describes OTEL settings loaded from env.
type Config struct {
	Enabled      bool
	Endpoint     string
	ServiceName  string
	ServiceVer   string
	Environment  string
	BatchTimeout time.Duration
}

// LoadConfigFromEnv reads tracing settings from environment.
func LoadConfigFromEnv() Config {
	enabled := strings.EqualFold(os.Getenv("ENABLE_TRACING"), "true")
	endpoint := os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT")
	if endpoint == "" {
		endpoint = "otel-collector:4318"
	}
	serviceName := os.Getenv("APP_NAME")
	if serviceName == "" {
		serviceName = "maxapp"
	}
	serviceVer := os.Getenv("APP_VERSION")
	if serviceVer == "" {
		serviceVer = "0.1.0"
	}
	env := os.Getenv("ENV")
	if env == "" {
		env = "development"
	}

	return Config{
		Enabled:      enabled,
		Endpoint:     endpoint,
		ServiceName:  serviceName,
		ServiceVer:   serviceVer,
		Environment:  env,
		BatchTimeout: 5 * time.Second,
	}
}

// Init initializes OTEL tracer provider and returns shutdown function.
func Init(ctx context.Context, cfg Config) (func(context.Context) error, error) {
	if !cfg.Enabled {
		// No-op shutdown when tracing is disabled.
		return func(context.Context) error { return nil }, nil
	}

	// Trim http(s) from endpoint for otlptracehttp.
	endpoint := trimScheme(cfg.Endpoint)
	exporter, err := otlptracehttp.New(ctx,
		otlptracehttp.WithEndpoint(endpoint),
		otlptracehttp.WithInsecure(),
	)
	if err != nil {
		return nil, err
	}

	// Attach common resource attributes for service discovery in Tempo/Grafana.
	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceNameKey.String(cfg.ServiceName),
			semconv.ServiceVersionKey.String(cfg.ServiceVer),
			semconv.DeploymentEnvironmentKey.String(cfg.Environment),
		),
	)
	if err != nil {
		return nil, err
	}

	// Use batch processor for production-friendly export.
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter, sdktrace.WithBatchTimeout(cfg.BatchTimeout)),
		sdktrace.WithResource(res),
	)

	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.TraceContext{})

	return tp.Shutdown, nil
}

// trimScheme removes http/https prefix from endpoint value.
func trimScheme(endpoint string) string {
	endpoint = strings.TrimPrefix(endpoint, "http://")
	endpoint = strings.TrimPrefix(endpoint, "https://")
	return endpoint
}
