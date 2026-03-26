package repository

import (
	"interfaces/internal/domain"
	"interfaces/internal/infrastructure/persistence"
)

type Repository struct {
	UserRepository
}

//go:generate mockery --name=UserRepository --filename=user_repository_mock.go --output=../mocks --case=underscore
type UserRepository interface {
	GetUserById(id int) (domain.User, error)
}

func NewRepository() *Repository {
	return &Repository{
		UserRepository: persistence.NewUserRepository(),
	}
}
