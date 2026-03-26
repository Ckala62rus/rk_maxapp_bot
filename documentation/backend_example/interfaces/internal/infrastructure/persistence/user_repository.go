package persistence

import (
	"interfaces/internal/domain"
	"time"
)

type UserRepository struct{}

func NewUserRepository() *UserRepository {
	return &UserRepository{}
}

func (r *UserRepository) GetUserById(id int) (domain.User, error) {
	return domain.User{
		ID:           1,
		Username:     "user",
		Email:        "user@mail.ru",
		PasswordHash: "123123",
		UpdatedAt:    time.Now(),
		CreatedAt:    time.Now(),
	}, nil
}
