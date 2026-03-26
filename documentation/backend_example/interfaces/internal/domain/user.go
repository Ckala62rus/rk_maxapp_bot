package domain

import "time"

type User struct {
	ID           int64     `json:"id" db:"id"`
	Username     string    `json:"name" db:"name"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"` // Не показываем в API Или `json:"password_hash,omitempty"` если нужно
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}
