package pkg

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// Client определяет интерфейс для работы с MinIO (S3-совместимое хранилище)
type Client interface {
	// ------------------- Upload методы -------------------

	// UploadFile загружает файл с локального диска в хранилище
	// bucket - имя корзины (бакета)
	// objectName - имя объекта в хранилище
	// filePath - путь к файлу на локальном диске
	UploadFile(ctx context.Context, bucket, objectName, filePath string) (minio.UploadInfo, error)

	// UploadStream загружает данные из потока (io.Reader) в хранилище
	// bucket - имя корзины
	// objectName - имя объекта в хранилище
	// reader - источник данных (например, тело HTTP-запроса, файл)
	// size - размер данных в байтах (если неизвестен, передайте -1)
	UploadStream(ctx context.Context, bucket, objectName string, reader io.Reader, size int64) (minio.UploadInfo, error)

	// UploadBytes загружает срез байт в хранилище
	// bucket - имя корзины
	// objectName - имя объекта в хранилище
	// data - срез байт для загрузки
	// contentType - MIME-тип (например, "image/jpeg", "application/json")
	UploadBytes(ctx context.Context, bucket, objectName string, data []byte, contentType string) (minio.UploadInfo, error)

	// ------------------- Download методы -------------------

	// GetObject возвращает объект из хранилища в виде потока (*minio.Object)
	// Важно: объект нужно закрыть после использования (defer obj.Close())
	// bucket - имя корзины
	// objectName - имя объекта в хранилище
	GetObject(ctx context.Context, bucket, objectName string) (*minio.Object, error)

	// DownloadFile скачивает объект из хранилища и сохраняет на локальный диск
	// bucket - имя корзины
	// objectName - имя объекта в хранилище
	// destPath - путь для сохранения файла на диске
	DownloadFile(ctx context.Context, bucket, objectName, destPath string) error

	// GetBytes скачивает объект из хранилища и возвращает его как срез байт
	// bucket - имя корзины
	// objectName - имя объекта в хранилище
	GetBytes(ctx context.Context, bucket, objectName string) ([]byte, error)

	// ------------------- URL методы (для прямого доступа) -------------------

	// GetPresignedURL генерирует временную ссылку для скачивания объекта
	// Ссылка действует ограниченное время (expiry) и позволяет скачать файл напрямую из MinIO
	// bucket - имя корзины
	// objectName - имя объекта в хранилище
	// expiry - время жизни ссылки (например, 15 * time.Minute)
	GetPresignedURL(ctx context.Context, bucket, objectName string, expiry time.Duration) (string, error)

	// GetPresignedUploadURL генерирует временную ссылку для загрузки объекта
	// Позволяет клиенту загрузить файл напрямую в MinIO минуя ваш бэкенд
	// bucket - имя корзины
	// objectName - имя объекта в хранилище
	// expiry - время жизни ссылки
	GetPresignedUploadURL(ctx context.Context, bucket, objectName string, expiry time.Duration) (string, error)

	// ------------------- Методы для получения информации -------------------

	// GetObjectInfo возвращает метаинформацию об объекте (размер, дата создания, Content-Type и т.д.)
	// bucket - имя корзины
	// objectName - имя объекта в хранилище
	GetObjectInfo(ctx context.Context, bucket, objectName string) (minio.ObjectInfo, error)

	// ListObjects возвращает канал с информацией об объектах в корзине
	// bucket - имя корзины
	// prefix - фильтр по префиксу (например, "images/")
	// recursive - true = показывать объекты во вложенных папках, false = только в текущей
	ListObjects(ctx context.Context, bucket, prefix string, recursive bool) (<-chan minio.ObjectInfo, error)

	// BucketExists проверяет, существует ли корзина с указанным именем
	// bucket - имя корзины
	BucketExists(ctx context.Context, bucket string) (bool, error)

	// ------------------- Методы удаления -------------------

	// DeleteObject удаляет один объект из корзины
	// bucket - имя корзины
	// objectName - имя объекта в хранилище
	DeleteObject(ctx context.Context, bucket, objectName string) error

	// DeleteObjects удаляет несколько объектов из корзины за один вызов
	// bucket - имя корзины
	// objectNames - срез имён объектов для удаления
	DeleteObjects(ctx context.Context, bucket string, objectNames []string) error

	// ------------------- Управление корзинами (бакетами) -------------------

	// MakeBucket создаёт новую корзину
	// bucket - имя новой корзины
	// opts - опции создания (например, регион)
	MakeBucket(ctx context.Context, bucket string, opts minio.MakeBucketOptions) error

	// RemoveBucket удаляет корзину (корзина должна быть пустой!)
	// bucket - имя корзины
	RemoveBucket(ctx context.Context, bucket string) error

	// ListBuckets возвращает список всех корзин
	ListBuckets(ctx context.Context) ([]minio.BucketInfo, error)

	// ------------------- Проверка доступности -------------------

	// HealthCheck проверяет доступность MinIO сервера
	// duration - таймаут ожидания ответа
	// возвращает true если сервер доступен, иначе ошибку
	HealthCheck(duration time.Duration) (bool, error)
}

// MinioClientConfig конфигурация для подключения к MinIO
type MinioClientConfig struct {
	Host     string
	User     string
	Password string
	SSL      bool
}

// MinioClient реализует интерфейс Client
type MinioClient struct {
	client *minio.Client
}

// NewMinioClient создаёт новый клиент MinIO и возвращает реализацию интерфейса
func NewMinioClient(config MinioClientConfig) (*MinioClient, error) {
	cli, err := minio.New(config.Host, &minio.Options{
		Creds:  credentials.NewStaticV4(config.User, config.Password, ""),
		Secure: config.SSL,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create minio client: %w", err)
	}

	return &MinioClient{client: cli}, nil
}

// MustMinioClient создаёт клиент и паникует при ошибке (для init)
func MustMinioClient(config MinioClientConfig) *MinioClient {
	client, err := NewMinioClient(config)
	if err != nil {
		panic(fmt.Sprintf("failed to create minio client: %v", err))
	}
	return client
}

// --- Upload methods ---

func (m *MinioClient) UploadFile(ctx context.Context, bucket, objectName, filePath string) (minio.UploadInfo, error) {
	info, err := m.client.FPutObject(ctx, bucket, objectName, filePath, minio.PutObjectOptions{})
	if err != nil {
		return minio.UploadInfo{}, fmt.Errorf("failed to upload file: %w", err)
	}
	return info, nil
}

func (m *MinioClient) UploadStream(ctx context.Context, bucket, objectName string, reader io.Reader, size int64) (minio.UploadInfo, error) {
	info, err := m.client.PutObject(ctx, bucket, objectName, reader, size, minio.PutObjectOptions{})
	if err != nil {
		return minio.UploadInfo{}, fmt.Errorf("failed to upload stream: %w", err)
	}
	return info, nil
}

func (m *MinioClient) UploadBytes(ctx context.Context, bucket, objectName string, data []byte, contentType string) (minio.UploadInfo, error) {
	reader := bytes.NewReader(data)
	info, err := m.client.PutObject(ctx, bucket, objectName, reader, int64(len(data)), minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return minio.UploadInfo{}, fmt.Errorf("failed to upload bytes: %w", err)
	}
	return info, nil
}

// --- Download methods ---

func (m *MinioClient) GetObject(ctx context.Context, bucket, objectName string) (*minio.Object, error) {
	obj, err := m.client.GetObject(ctx, bucket, objectName, minio.GetObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get object: %w", err)
	}
	return obj, nil
}

func (m *MinioClient) DownloadFile(ctx context.Context, bucket, objectName, destPath string) error {
	err := m.client.FGetObject(ctx, bucket, objectName, destPath, minio.GetObjectOptions{})
	if err != nil {
		return fmt.Errorf("failed to download file: %w", err)
	}
	return nil
}

func (m *MinioClient) GetBytes(ctx context.Context, bucket, objectName string) ([]byte, error) {
	obj, err := m.GetObject(ctx, bucket, objectName)
	if err != nil {
		return nil, err
	}
	defer obj.Close()

	data, err := io.ReadAll(obj)
	if err != nil {
		return nil, fmt.Errorf("failed to read object data: %w", err)
	}
	return data, nil
}

// --- URL methods ---

func (m *MinioClient) GetPresignedURL(ctx context.Context, bucket, objectName string, expiry time.Duration) (string, error) {
	url, err := m.client.PresignedGetObject(ctx, bucket, objectName, expiry, nil)
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned URL: %w", err)
	}
	return url.String(), nil
}

func (m *MinioClient) GetPresignedUploadURL(ctx context.Context, bucket, objectName string, expiry time.Duration) (string, error) {
	url, err := m.client.PresignedPutObject(ctx, bucket, objectName, expiry)
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned upload URL: %w", err)
	}
	return url.String(), nil
}

// --- Metadata methods ---

func (m *MinioClient) GetObjectInfo(ctx context.Context, bucket, objectName string) (minio.ObjectInfo, error) {
	info, err := m.client.StatObject(ctx, bucket, objectName, minio.StatObjectOptions{})
	if err != nil {
		return minio.ObjectInfo{}, fmt.Errorf("failed to get object info: %w", err)
	}
	return info, nil
}

func (m *MinioClient) ListObjects(ctx context.Context, bucket, prefix string, recursive bool) (<-chan minio.ObjectInfo, error) {
	// Проверяем существование bucket'а
	exists, err := m.BucketExists(ctx, bucket)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, fmt.Errorf("bucket %s does not exist", bucket)
	}

	objectCh := m.client.ListObjects(ctx, bucket, minio.ListObjectsOptions{
		Prefix:    prefix,
		Recursive: recursive,
	})
	return objectCh, nil
}

func (m *MinioClient) BucketExists(ctx context.Context, bucket string) (bool, error) {
	exists, err := m.client.BucketExists(ctx, bucket)
	if err != nil {
		return false, fmt.Errorf("failed to check bucket existence: %w", err)
	}
	return exists, nil
}

// --- Delete methods ---

func (m *MinioClient) DeleteObject(ctx context.Context, bucket, objectName string) error {
	err := m.client.RemoveObject(ctx, bucket, objectName, minio.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete object: %w", err)
	}
	return nil
}

func (m *MinioClient) DeleteObjects(ctx context.Context, bucket string, objectNames []string) error {
	objectsCh := make(chan minio.ObjectInfo)
	go func() {
		defer close(objectsCh)
		for _, name := range objectNames {
			objectsCh <- minio.ObjectInfo{Key: name}
		}
	}()

	for err := range m.client.RemoveObjects(ctx, bucket, objectsCh, minio.RemoveObjectsOptions{}) {
		if err.Err != nil {
			return fmt.Errorf("failed to delete objects: %w", err.Err)
		}
	}
	return nil
}

// --- Bucket management ---

func (m *MinioClient) MakeBucket(ctx context.Context, bucket string, opts minio.MakeBucketOptions) error {
	err := m.client.MakeBucket(ctx, bucket, opts)
	if err != nil {
		return fmt.Errorf("failed to create bucket: %w", err)
	}
	return nil
}

func (m *MinioClient) RemoveBucket(ctx context.Context, bucket string) error {
	err := m.client.RemoveBucket(ctx, bucket)
	if err != nil {
		return fmt.Errorf("failed to remove bucket: %w", err)
	}
	return nil
}

func (m *MinioClient) ListBuckets(ctx context.Context) ([]minio.BucketInfo, error) {
	buckets, err := m.client.ListBuckets(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list buckets: %w", err)
	}
	return buckets, nil
}

// --- Health ---

func (m *MinioClient) HealthCheck(duration time.Duration) (bool, error) {
	_, err := m.client.HealthCheck(duration)
	if err != nil {
		return false, fmt.Errorf("health check failed: %w", err)
	}
	return true, nil
}
