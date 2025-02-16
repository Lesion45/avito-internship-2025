package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"avito-internship/internal/entity"
	"avito-internship/internal/repository"
	"avito-internship/internal/repository/repoerrs"
	"avito-internship/internal/service/servicerrs"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

func TestAuthService_Authorization(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := repository.NewMockUser(ctrl)
	logger := zap.NewNop()
	tokenTTL := time.Hour
	salt := "test-salt"

	service := NewAuthService(logger, mockRepo, tokenTTL, salt)

	tests := []struct {
		name          string
		username      string
		password      string
		mockRepoSetup func()
		expectedToken string
		expectedError error
	}{
		{
			name:     "Successful authorization",
			username: "user1",
			password: "password123",
			mockRepoSetup: func() {
				hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
				mockRepo.EXPECT().
					GetUserCredentials(gomock.Any(), "user1").
					Return(entity.User{Password: hashedPassword}, nil)
			},
			expectedToken: "valid-token",
			expectedError: nil,
		},
		{
			name:     "User not found, successful registration",
			username: "user1",
			password: "password123",
			mockRepoSetup: func() {
				mockRepo.EXPECT().
					GetUserCredentials(gomock.Any(), "user1").
					Return(entity.User{}, repoerrs.ErrUserNotFound)
				mockRepo.EXPECT().
					AddUser(gomock.Any(), "user1", gomock.Any()).
					Return(nil)
			},
			expectedToken: "valid-token",
			expectedError: nil,
		},
		{
			name:     "Invalid credentials",
			username: "user1",
			password: "wrongpassword",
			mockRepoSetup: func() {
				hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
				mockRepo.EXPECT().
					GetUserCredentials(gomock.Any(), "user1").
					Return(entity.User{Password: hashedPassword}, nil)
			},
			expectedToken: "",
			expectedError: servicerrs.ErrInvalidCredentials,
		},
		{
			name:     "Repository error",
			username: "user1",
			password: "password123",
			mockRepoSetup: func() {
				mockRepo.EXPECT().
					GetUserCredentials(gomock.Any(), "user1").
					Return(entity.User{}, errors.New("repository error"))
			},
			expectedToken: "",
			expectedError: errors.New("repository error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockRepoSetup()

			token, err := service.Authorization(context.Background(), tt.username, tt.password)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError.Error())
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, token)
			}
		})
	}
}
