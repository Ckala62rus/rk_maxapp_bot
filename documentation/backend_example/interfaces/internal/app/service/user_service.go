package internal

import (
	"interfaces/internal/domain"
	"interfaces/internal/domain/repository"
)

type UserService struct {
	reps *repository.Repository
}

func NewUserService(reps *repository.Repository) *UserService {
	return &UserService{reps: reps}
}

func (r *UserService) GetUserById(id int) (domain.User, error) {
	return r.reps.GetUserById(1)
}
