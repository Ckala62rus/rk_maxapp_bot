package service

import (
	internal "interfaces/internal/app/service"
	"interfaces/internal/domain"
	"interfaces/internal/domain/repository"
)

type Service struct {
	UserService
}

type UserService interface {
	GetUserById(id int) (domain.User, error)
}

func NewService(reps *repository.Repository) *Service {
	return &Service{
		UserService: internal.NewUserService(reps),
	}
}
