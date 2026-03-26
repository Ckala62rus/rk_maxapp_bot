// Package service contains business logic.
package service

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"log/slog"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/Ckala62rus/rk_maxapp_bot/backend/internal/domain"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

// ErrInvalidInitData indicates invalid MAX initData payload.
var ErrInvalidInitData = errors.New("invalid init data")

// AuthService validates MAX initData and ensures user exists.
type AuthService struct {
	userRepo domain.UserRepository
	botToken string
	logger   *slog.Logger
}

// MaxUser is minimal parsed structure from initData user JSON.
type MaxUser struct {
	ID           int64   `json:"id"`
	FirstName    string  `json:"first_name"`
	LastName     string  `json:"last_name"`
	Username     *string `json:"username"`
	LanguageCode string  `json:"language_code"`
	PhotoURL     *string `json:"photo_url"`
}

// InitDataInput describes mock user data for dev initData generation.
type InitDataInput struct {
	UserID       int64
	FirstName    string
	LastName     string
	Username     string
	LanguageCode string
	PhotoURL     string
}

// NewAuthService constructs AuthService with repo and bot token.
func NewAuthService(userRepo domain.UserRepository, botToken string, logger *slog.Logger) *AuthService {
	return &AuthService{
		userRepo: userRepo,
		botToken: botToken,
		logger:   logger,
	}
}

// ValidateInitData verifies signature and returns user record.
//
// Подробно, как работает авторизация через initData (MAX WebApp):
// 1) Клиент получает строку initData (URL-encoded query string) от WebApp SDK и
//    передает ее либо в POST /api/auth/validate (body.initData), либо в заголовке
//    X-Max-InitData для всех защищенных запросов.
// 2) Мы делаем URL-decode строки и парсим ее как query-параметры (key=value).
// 3) Достаем параметр hash и строим data-check string: сортируем ключи,
//    исключая hash, и склеиваем строки вида key=value через "\n".
// 4) Вычисляем секрет: HMAC-SHA256("WebAppData", botToken) и затем ожидаемый hash:
//    HMAC-SHA256(secret, data-check string). Сравниваем с полученным hash.
// 5) Если подпись корректна, парсим параметр user (JSON) и берем max user id.
// 6) По max user id ищем пользователя в БД, если нет — создаем с флагами
//    IsAdmin/IsBlocked/IsApproved = false.
// 7) Возвращаем пользователя и его флаги — UI использует их для доступа.
func (s *AuthService) ValidateInitData(ctx context.Context, initData string) (*domain.User, error) {
	tracer := otel.Tracer("service.auth")
	ctx, span := tracer.Start(ctx, "ValidateInitData")
	defer span.End()

	// Validate signature and extract MAX user id.
	maxUser, err := s.validateSignature(initData)
	if err != nil {
		span.SetAttributes(attribute.String("auth.error", err.Error()))
		s.logger.Debug("initData validation failed", "error", err)
		return nil, err
	}
	span.SetAttributes(attribute.Int64("max_user_id", maxUser.ID))

	// Try to find existing user.
	user, err := s.userRepo.GetByMaxUserID(ctx, maxUser.ID)
	if err == nil {
		return user, nil
	}
	if !errors.Is(err, domain.ErrNotFound) {
		return nil, err
	}

	// Create new user on first login.
	newUser := domain.User{
		MaxUserID:    maxUser.ID,
		MaxUsername:  valueOrEmpty(maxUser.Username),
		MaxFirstName: maxUser.FirstName,
		MaxLastName:  maxUser.LastName,
		LanguageCode: maxUser.LanguageCode,
		PhotoURL:     valueOrEmpty(maxUser.PhotoURL),
		FirstName:    "",
		LastName:     "",
		IsAdmin:      false,
		IsBlocked:    false,
		IsApproved:   false,
	}

	created, err := s.userRepo.Create(ctx, newUser)
	if err != nil {
		s.logger.Error("failed to create user", "error", err)
		return nil, err
	}
	return created, nil
}

// validateSignature verifies HMAC signature and returns MAX user payload.
func (s *AuthService) validateSignature(initData string) (MaxUser, error) {
	if initData == "" {
		return MaxUser{}, ErrInvalidInitData
	}

	// initData is URL-encoded, decode it first.
	decoded, err := url.QueryUnescape(initData)
	if err != nil {
		return MaxUser{}, ErrInvalidInitData
	}

	// Parse query string to key/value map.
	values, err := url.ParseQuery(decoded)
	if err != nil {
		return MaxUser{}, ErrInvalidInitData
	}

	// Extract hash and validate it.
	hash := values.Get("hash")
	if hash == "" {
		return MaxUser{}, ErrInvalidInitData
	}

	// Build data-check string according to MAX spec.
	dataCheck := buildDataCheckString(values)
	secret := hmacSHA256([]byte("WebAppData"), []byte(s.botToken))
	expected := hmacSHA256(secret, []byte(dataCheck))

	expectedHex := hex.EncodeToString(expected)
	if !strings.EqualFold(expectedHex, hash) {
		return MaxUser{}, ErrInvalidInitData
	}

	// Parse user json to get MAX user id.
	userJSON := values.Get("user")
	if userJSON == "" {
		return MaxUser{}, ErrInvalidInitData
	}

	var maxUser MaxUser
	if err := json.Unmarshal([]byte(userJSON), &maxUser); err != nil {
		return MaxUser{}, ErrInvalidInitData
	}

	if maxUser.ID == 0 {
		return MaxUser{}, ErrInvalidInitData
	}

	return maxUser, nil
}

// GenerateInitData builds signed initData for local testing.
func (s *AuthService) GenerateInitData(ctx context.Context, input InitDataInput) (string, error) {
	tracer := otel.Tracer("service.auth")
	ctx, span := tracer.Start(ctx, "GenerateInitData")
	defer span.End()

	// Apply defaults for easier testing.
	if input.UserID == 0 {
		input.UserID = 1001
	}
	if input.FirstName == "" {
		input.FirstName = "Test"
	}
	if input.LastName == "" {
		input.LastName = "User"
	}
	if input.LanguageCode == "" {
		input.LanguageCode = "ru"
	}

	// Prepare user object with nullable fields.
	user := struct {
		ID           int64   `json:"id"`
		FirstName    string  `json:"first_name"`
		LastName     string  `json:"last_name"`
		Username     *string `json:"username"`
		LanguageCode string  `json:"language_code"`
		PhotoURL     *string `json:"photo_url"`
	}{
		ID:           input.UserID,
		FirstName:    input.FirstName,
		LastName:     input.LastName,
		Username:     nilIfEmpty(input.Username),
		LanguageCode: input.LanguageCode,
		PhotoURL:     nilIfEmpty(input.PhotoURL),
	}

	userJSON, err := json.Marshal(user)
	if err != nil {
		return "", err
	}

	// Build base values without hash first.
	values := url.Values{}
	values.Set("auth_date", strconv.FormatInt(time.Now().Unix(), 10))
	values.Set("query_id", uuid.NewString())
	values.Set("user", string(userJSON))

	// Compute hash according to MAX spec.
	dataCheck := buildDataCheckString(values)
	secret := hmacSHA256([]byte("WebAppData"), []byte(s.botToken))
	hash := hex.EncodeToString(hmacSHA256(secret, []byte(dataCheck)))
	values.Set("hash", hash)

	// Encode to URL-style query string.
	return values.Encode(), nil
}

// buildDataCheckString sorts keys and joins key=value by newlines.
func buildDataCheckString(values url.Values) string {
	keys := make([]string, 0, len(values))
	for key := range values {
		if key == "hash" {
			continue
		}
		keys = append(keys, key)
	}
	sort.Strings(keys)

	pairs := make([]string, 0, len(keys))
	for _, key := range keys {
		pairs = append(pairs, key+"="+values.Get(key))
	}
	return strings.Join(pairs, "\n")
}

// hmacSHA256 computes HMAC-SHA256.
func hmacSHA256(key, data []byte) []byte {
	mac := hmac.New(sha256.New, key)
	_, _ = mac.Write(data)
	return mac.Sum(nil)
}

// valueOrEmpty converts nil string pointers to empty string.
func valueOrEmpty(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

// nilIfEmpty converts empty string to nil.
func nilIfEmpty(value string) *string {
	if value == "" {
		return nil
	}
	return &value
}
