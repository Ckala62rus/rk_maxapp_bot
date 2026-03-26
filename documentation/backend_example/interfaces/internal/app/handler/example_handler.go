package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/minio/minio-go/v7"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"io"
	"net/http"
	"path/filepath"
	"time"
)

func (h *Handler) SayHello(w http.ResponseWriter, r *http.Request) {
	a := "**** 000 123 000 ****"
	spanCtx := trace.SpanContextFromContext(r.Context())
	h.log.Info(
		"Say Hello handler",
		"message", a,
		"method", r.Method,
		"path", r.URL.Path,
		"remote", r.RemoteAddr,
		"trace_id", spanCtx.TraceID().String(),
		"span_id", spanCtx.SpanID().String(),
	)
	w.Write([]byte("example handler"))
}

func (h *Handler) TraceTest(w http.ResponseWriter, r *http.Request) {
	ctx, span := otel.Tracer("handler").Start(r.Context(), "TraceTest")
	defer span.End()

	span.SetAttributes(
		attribute.String("demo.trace", "true"),
	)

	h.log.Info("Trace test handler")
	w.Write([]byte("trace test"))
	_ = ctx
}

func (h *Handler) GetFileByName(w http.ResponseWriter, r *http.Request) {
	h.log.Info("Get file by name handler")
	fileName := r.URL.Query().Get("file")
	if fileName == "" {
		http.Error(w, "file parameter is required", http.StatusBadRequest)
		return
	}

	// 3. Проверяем, существует ли файл в MinIO
	ctx := context.Background()
	bucket := "images" // ваш bucket

	// Получаем информацию о файле
	info, err := h.minio.GetObjectInfo(ctx, bucket, fileName)
	if err != nil {
		h.log.Error("file not found: " + fileName + ", error: " + err.Error())
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}

	// 4. Генерируем временную ссылку для скачивания
	presignedURL, err := h.minio.GetPresignedURL(ctx, bucket, fileName, 15*time.Minute)
	if err != nil {
		h.log.Error("failed to generate presigned URL: " + err.Error())
		http.Error(w, "Failed to generate download link", http.StatusInternalServerError)
		return
	}

	// 5. Возвращаем информацию о файле
	response := map[string]interface{}{
		"file_name":     fileName,
		"size":          info.Size,
		"content_type":  info.ContentType,
		"last_modified": info.LastModified,
		"download_url":  presignedURL,
		"expires_in":    "15 minutes",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *Handler) UploadFile(w http.ResponseWriter, r *http.Request) {
	h.log.Info("Upload file handler")
	file, header, err := r.FormFile("file")
	defer file.Close()
	if err != nil {
		h.log.Error("error save file" + err.Error())
	}

	fmt.Printf("Received file: %s, size: %d, content-type: %s\n",
		header.Filename, header.Size, header.Header.Get("Content-Type"))

	minioClient := h.minio

	// Используем
	ctx := context.Background()

	fileBytes, err := io.ReadAll(file)
	if err != nil {
		h.log.Error("failed to read file: " + err.Error())
		http.Error(w, "Failed to read file", http.StatusInternalServerError)
		return
	}

	isExists, err := minioClient.BucketExists(ctx, "images")
	if err != nil {
		h.log.Error("bucket check exist " + err.Error())
		http.Error(w, "bucket check exist", http.StatusInternalServerError)
	}

	if !isExists {
		err = minioClient.MakeBucket(ctx, "images", minio.MakeBucketOptions{})
		if err != nil {
			h.log.Error("bucket create " + err.Error())
			http.Error(w, "bucket create", http.StatusInternalServerError)
		}
	}

	info, err := minioClient.UploadBytes(ctx, "images", header.Filename, fileBytes, header.Header.Get("Content-Type"))
	if err != nil {
		h.log.Error("failed to upload to minio: " + err.Error())
		http.Error(w, "Failed to upload file", http.StatusInternalServerError)
		return
	}

	fmt.Printf("File uploaded successfully, ETag: %s\n", info.ETag)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("File uploaded successfully"))
}

func (h *Handler) DownloadFile(w http.ResponseWriter, r *http.Request) {
	// 1. Получаем имя файла из query-параметра
	fileName := r.URL.Query().Get("file")
	if fileName == "" {
		h.log.Error("file parameter is missing")
		http.Error(w, "file parameter is required", http.StatusBadRequest)
		return
	}

	ctx := context.Background()
	bucket := "images" // ваш bucket

	// 2. Получаем объект из MinIO через GetObject
	// GetObject возвращает *minio.Object, который реализует io.ReadCloser
	obj, err := h.minio.GetObject(ctx, bucket, fileName)
	if err != nil {
		h.log.Error(fmt.Sprintf("failed to get object: %v", err))
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}
	// ВАЖНО: всегда закрываем объект после использования!
	defer obj.Close()

	// 3. Получаем информацию о файле (размер, content-type и т.д.)
	// Stat() возвращает minio.ObjectInfo
	objInfo, err := obj.Stat()
	if err != nil {
		h.log.Error(fmt.Sprintf("failed to get object info: %v", err))
		http.Error(w, "Failed to get file info", http.StatusInternalServerError)
		return
	}

	// 4. Устанавливаем правильные заголовки для скачивания
	// Content-Type из MinIO (или определяем по расширению)
	contentType := objInfo.ContentType
	if contentType == "" {
		// Если MinIO не определил тип, пробуем по расширению
		ext := filepath.Ext(fileName)
		switch ext {
		case ".jpg", ".jpeg":
			contentType = "image/jpeg"
		case ".png":
			contentType = "image/png"
		case ".gif":
			contentType = "image/gif"
		case ".pdf":
			contentType = "application/pdf"
		default:
			contentType = "application/octet-stream"
		}
	}

	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", fileName))
	w.Header().Set("Content-Length", fmt.Sprintf("%d", objInfo.Size))

	// 5. Копируем содержимое файла напрямую в http.ResponseWriter
	// Это эффективно, данные идут потоком без загрузки в память целиком
	bytesWritten, err := io.Copy(w, obj)
	if err != nil {
		h.log.Error(fmt.Sprintf("failed to send file: %v", err))
		// Не отправляем ошибку через http.Error, так как заголовки уже отправлены
		return
	}

	h.log.Info(fmt.Sprintf("file %s downloaded successfully, size: %d bytes",
		fileName, bytesWritten))
}
