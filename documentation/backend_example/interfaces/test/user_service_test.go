package test

import (
	"github.com/stretchr/testify/assert"
	"interfaces/internal/domain"
	"interfaces/internal/domain/mocks"
	"interfaces/internal/domain/repository"
	"interfaces/internal/domain/service"
	"testing"
)

func TestUserService_GetUserByID(t *testing.T) {
	t.Run("success - user found", func(t *testing.T) {
		// Arrange
		mockRepo := &mocks.UserRepository{}

		// Создаем реальную структуру Repository с мокнутым UserRepository внутри
		repo := &repository.Repository{
			UserRepository: mockRepo, // Встраиваем мок
		}

		services := service.NewService(repo)

		expectedUser := domain.User{
			ID:       1,
			Username: "John Doe",
			Email:    "john@example.com",
		}

		// Настраиваем мок
		mockRepo.On("GetUserById", 1).
			Return(expectedUser, nil).
			Once()

		// Act
		user, err := services.UserService.GetUserById(1)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, expectedUser, user)
		mockRepo.AssertExpectations(t)
	})
}
