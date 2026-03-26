package test

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/stretchr/testify/require"
)

func TestMinIOSmoke_ThroughNginx(t *testing.T) {
	t.Parallel()

	endpoint := envOr("MINIO_ENDPOINT", "localhost:9000")
	accessKey := envOr("MINIO_ROOT_USER", "minioadmin")
	secretKey := envOr("MINIO_ROOT_PASSWORD", "minioadmin")
	useSSL := strings.EqualFold(envOr("MINIO_USE_SSL", "false"), "true")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: useSSL,
	})
	require.NoError(t, err)

	_, err = client.HealthCheck(2 * time.Second)
	if err != nil {
		t.Skipf("skip smoke test: MinIO is unavailable at %s (%v)", endpoint, err)
	}

	bucketName := "smoke-bucket-" + time.Now().Format("20060102150405")
	objectName := "hello.txt"
	content := []byte("minio smoke test via nginx")

	err = client.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{})
	require.NoError(t, err)
	t.Cleanup(func() {
		_ = client.RemoveObject(context.Background(), bucketName, objectName, minio.RemoveObjectOptions{})
		_ = client.RemoveBucket(context.Background(), bucketName)
	})

	_, err = client.PutObject(ctx, bucketName, objectName, bytes.NewReader(content), int64(len(content)), minio.PutObjectOptions{
		ContentType: "text/plain",
	})
	require.NoError(t, err)

	presignedURL, err := client.PresignedGetObject(ctx, bucketName, objectName, 10*time.Minute, nil)
	require.NoError(t, err)
	require.Contains(t, presignedURL.Host, endpoint)

	httpClient := &http.Client{Timeout: 10 * time.Second}
	resp, err := httpClient.Get(presignedURL.String())
	require.NoError(t, err)
	defer func() {
		_ = resp.Body.Close()
	}()
	require.Equal(t, http.StatusOK, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	require.Equal(t, content, body)
}

func envOr(key, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}
