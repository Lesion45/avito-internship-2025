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

func TestOperationService_TransferFunds(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := repository.NewMockOperation(ctrl)
	mockCache := cache.NewMockCache(ctrl)
	logger := zap.NewNop()

	service := NewOperationService(logger, mockCache, mockRepo)

	tests := []struct {
		name           string
		input          TransferFundsInput
		mockRepoSetup  func(*repository.MockOperation)
		mockCacheSetup func(*cache.MockCache)
		expectedError  error
	}{
		{
			name: "Successful transfer",
			input: TransferFundsInput{
				Sender:    "user1",
				Recipient: "user2",
				Amount:    100,
			},
			mockRepoSetup: func(m *repository.MockOperation) {
				m.EXPECT().
					SaveTransfer(gomock.Any(), "user1", "user2", 100).
					Return(nil)
			},
			mockCacheSetup: func(m *cache.MockCache) {
				m.EXPECT().
					Del(gomock.Any(), "user_info:user1", "user_info:user2").
					Return(nil)
			},
			expectedError: nil,
		},
		{
			name: "Recipient not found",
			input: TransferFundsInput{
				Sender:    "user1",
				Recipient: "user2",
				Amount:    100,
			},
			mockRepoSetup: func(m *repository.MockOperation) {
				m.EXPECT().
					SaveTransfer(gomock.Any(), "user1", "user2", 100).
					Return(repoerrs.ErrUserNotFound)
			},
			mockCacheSetup: func(m *cache.MockCache) {},
			expectedError:  servicerrs.ErrRecipientNotFound,
		},
		{
			name: "Insufficient funds",
			input: TransferFundsInput{
				Sender:    "user1",
				Recipient: "user2",
				Amount:    100,
			},
			mockRepoSetup: func(m *repository.MockOperation) {
				m.EXPECT().
					SaveTransfer(gomock.Any(), "user1", "user2", 100).
					Return(repoerrs.ErrInsufficientFunds)
			},
			mockCacheSetup: func(m *cache.MockCache) {},
			expectedError:  servicerrs.ErrInsufficientFunds,
		},
		{
			name: "Repository error",
			input: TransferFundsInput{
				Sender:    "user1",
				Recipient: "user2",
				Amount:    100,
			},
			mockRepoSetup: func(m *repository.MockOperation) {
				m.EXPECT().
					SaveTransfer(gomock.Any(), "user1", "user2", 100).
					Return(errors.New("repository error"))
			},
			mockCacheSetup: func(m *cache.MockCache) {},
			expectedError:  errors.New("repository error"),
		},
		{
			name: "Cache invalidation error",
			input: TransferFundsInput{
				Sender:    "user1",
				Recipient: "user2",
				Amount:    100,
			},
			mockRepoSetup: func(m *repository.MockOperation) {
				m.EXPECT().
					SaveTransfer(gomock.Any(), "user1", "user2", 100).
					Return(nil)
			},
			mockCacheSetup: func(m *cache.MockCache) {
				m.EXPECT().
					Del(gomock.Any(), "user_info:user1", "user_info:user2").
					Return(errors.New("cache error"))
			},
			expectedError: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockRepoSetup(mockRepo)
			tt.mockCacheSetup(mockCache)

			err := service.TransferFunds(context.Background(), tt.input)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestOperationService_PurchaseProduct(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := repository.NewMockOperation(ctrl)
	mockCache := cache.NewMockCache(ctrl)
	logger := zap.NewNop()

	service := NewOperationService(logger, mockCache, mockRepo)

	tests := []struct {
		name           string
		input          PurchaseProductInput
		mockRepoSetup  func(*repository.MockOperation)
		mockCacheSetup func(*cache.MockCache)
		expectedError  error
	}{
		{
			name: "Successful purchase",
			input: PurchaseProductInput{
				Username: "user1",
				Product:  "product1",
			},
			mockRepoSetup: func(m *repository.MockOperation) {
				m.EXPECT().
					SavePurchase(gomock.Any(), "user1", "product1").
					Return(nil)
			},
			mockCacheSetup: func(m *cache.MockCache) {
				m.EXPECT().
					Del(gomock.Any(), "user_info:user1").
					Return(nil)
			},
			expectedError: nil,
		},
		{
			name: "Customer not found",
			input: PurchaseProductInput{
				Username: "user1",
				Product:  "product1",
			},
			mockRepoSetup: func(m *repository.MockOperation) {
				m.EXPECT().
					SavePurchase(gomock.Any(), "user1", "product1").
					Return(repoerrs.ErrUserNotFound)
			},
			mockCacheSetup: func(m *cache.MockCache) {},
			expectedError:  servicerrs.ErrCustomerNotFound,
		},
		{
			name: "Product not found",
			input: PurchaseProductInput{
				Username: "user1",
				Product:  "product1",
			},
			mockRepoSetup: func(m *repository.MockOperation) {
				m.EXPECT().
					SavePurchase(gomock.Any(), "user1", "product1").
					Return(repoerrs.ErrProductNotFound)
			},
			mockCacheSetup: func(m *cache.MockCache) {},
			expectedError:  servicerrs.ErrProductNotFound,
		},
		{
			name: "Insufficient funds",
			input: PurchaseProductInput{
				Username: "user1",
				Product:  "product1",
			},
			mockRepoSetup: func(m *repository.MockOperation) {
				m.EXPECT().
					SavePurchase(gomock.Any(), "user1", "product1").
					Return(repoerrs.ErrInsufficientFunds)
			},
			mockCacheSetup: func(m *cache.MockCache) {},
			expectedError:  servicerrs.ErrInsufficientFunds,
		},
		{
			name: "Cache invalidation error",
			input: PurchaseProductInput{
				Username: "user1",
				Product:  "product1",
			},
			mockRepoSetup: func(m *repository.MockOperation) {
				m.EXPECT().
					SavePurchase(gomock.Any(), "user1", "product1").
					Return(nil)
			},
			mockCacheSetup: func(m *cache.MockCache) {
				m.EXPECT().
					Del(gomock.Any(), "user_info:user1").
					Return(errors.New("cache error"))
			},
			expectedError: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockRepoSetup(mockRepo)
			tt.mockCacheSetup(mockCache)

			err := service.PurchaseProduct(context.Background(), tt.input)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
