package handler

import (
	"log/slog"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	custommiddleware "interfaces/internal/app/middleware"
	"interfaces/internal/domain/service"
	"interfaces/pkg"
)

type Handler struct {
	services *service.Service
	log      *slog.Logger
	minio    *pkg.MinioClient
}

func NewHandler(services *service.Service, log *slog.Logger, minio *pkg.MinioClient) *Handler {
	return &Handler{
		services: services,
		log:      log,
		minio:    minio,
	}
}

func (h *Handler) InitRoutes() *chi.Mux {
	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(custommiddleware.CORSMiddleware())
	r.Use(custommiddleware.MetricsMiddleware())

	r.Handle("/metrics", promhttp.Handler())

	r.Route("/api", func(r chi.Router) {
		r.Get("/", h.SayHello)
		r.Get("/trace-test", h.TraceTest)

		r.Post("/file/upload", h.UploadFile)
		r.Get("/file/get-file-presigned", h.GetFileByName)
		r.Get("/file/get-file-by-name", h.DownloadFile)
	})

	return r
}
