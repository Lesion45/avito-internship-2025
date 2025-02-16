package service

import (
	"context"
	"errors"
	"testing"

	"avito-internship/internal/cache"
	"avito-internship/internal/repository"
	"avito-internship/internal/repository/repoerrs"
	"avito-internship/internal/service/servicerrs"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestUserService_CreateUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := repository.NewMockUser(ctrl)
	mockCache := cache.NewMockCache(ctrl)
	logger := zap.NewNop()

	service := NewUserService(logger, mockCache, mockRepo)

	tests := []struct {
		name           string
		input          UserCreateInput
		mockRepoSetup  func(*repository.MockUser)
		mockCacheSetup func(*cache.MockCache)
		expectedError  error
	}{
		{
			name: "Success: user created",
			input: UserCreateInput{
				Username: "testuser",
				Password: []byte("password123"),
			},
			mockRepoSetup: func(m *repository.MockUser) {
				m.EXPECT().
					AddUser(gomock.Any(), "testuser", []byte("password123")).
					Return(nil)
			},
			mockCacheSetup: func(m *cache.MockCache) {},
			expectedError:  nil,
		},
		{
			name: "Error: user already exists",
			input: UserCreateInput{
				Username: "testuser",
				Password: []byte("password123"),
			},
			mockRepoSetup: func(m *repository.MockUser) {
				m.EXPECT().
					AddUser(gomock.Any(), "testuser", []byte("password123")).
					Return(repoerrs.ErrUserAlreadyExists)
			},
			mockCacheSetup: func(m *cache.MockCache) {},
			expectedError:  servicerrs.ErrUserAlreadyExists,
		},
		{
			name: "Error: repository error",
			input: UserCreateInput{
				Username: "testuser",
				Password: []byte("password123"),
			},
			mockRepoSetup: func(m *repository.MockUser) {
				m.EXPECT().
					AddUser(gomock.Any(), "testuser", []byte("password123")).
					Return(errors.New("repository error"))
			},
			mockCacheSetup: func(m *cache.MockCache) {},
			expectedError:  errors.New("repository error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockRepoSetup(mockRepo)
			tt.mockCacheSetup(mockCache)

			err := service.CreateUser(context.Background(), tt.input)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
