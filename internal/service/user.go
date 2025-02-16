package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"avito-internship/internal/cache"
	"avito-internship/internal/entity"
	"avito-internship/internal/repository"
	"avito-internship/internal/repository/repoerrs"
	"avito-internship/internal/service/servicerrs"

	"go.uber.org/zap"
)

type UserService struct {
	Log   *zap.Logger
	Cache cache.Cache
	Repo  repository.User
}

func NewUserService(log *zap.Logger, cache cache.Cache, repo repository.User) *UserService {
	return &UserService{
		Log:   log,
		Cache: cache,
		Repo:  repo,
	}
}

func (s *UserService) CreateUser(ctx context.Context, input UserCreateInput) error {
	const op = "service.UserService.CreateUser"

	s.Log.Info("Attempting to create user")

	err := s.Repo.AddUser(ctx, input.Username, input.Password)
	if err != nil {
		if errors.Is(err, repoerrs.ErrUserAlreadyExists) {
			s.Log.Warn("User already exists",
				zap.String("op", op),
				zap.String("username", input.Username),
			)

			return fmt.Errorf("%s: %w", op, servicerrs.ErrUserAlreadyExists)
		}

		s.Log.Error("Failed to create user ",
			zap.String("op", op),
			zap.Error(err),
		)

		return fmt.Errorf("%s: %w", op, err)
	}

	s.Log.Info("User successfully created")

	return nil
}

func (s *UserService) RetrieveUserInfo(ctx context.Context, input RetrieveUserInfoInput) (RetrieveUserInfoOutput, error) {
	const op = "service.UserService.RetrieveUserInfo"

	s.Log.Info("Attempting to retrieve info about user")

	cacheKey := fmt.Sprintf("user_info:%s", input.Username)

	var output RetrieveUserInfoOutput
	cachedData, err := s.Cache.Get(ctx, cacheKey)
	if err == nil {
		if err := json.Unmarshal([]byte(cachedData), &output); err == nil {
			s.Log.Info("Retrieved user info from cache", zap.String("username", input.Username))
			return output, nil
		}
	}

	balance, operations, inventory, err := s.Repo.GetInfo(ctx, input.Username)
	if err != nil {
		if errors.Is(err, repoerrs.ErrUserNotFound) {
			s.Log.Warn("User not found",
				zap.String("op", op),
				zap.String("username", input.Username),
			)

			return RetrieveUserInfoOutput{}, fmt.Errorf("%s: %w", op, servicerrs.ErrUserNotFound)
		}

		s.Log.Error("Failed to retrieve user info",
			zap.String("op", op),
			zap.Error(err),
		)

		return RetrieveUserInfoOutput{}, fmt.Errorf("%s: %w", op, err)
	}

	var transferIn []entity.Transfer
	var transferOut []entity.Transfer
	for _, operation := range operations {
		if operation.User == input.Username {
			transfer := entity.Transfer{
				Username: operation.Counterparty,
				Amount:   operation.Amount,
			}
			transferOut = append(transferOut, transfer)
		} else {
			transfer := entity.Transfer{
				Username: operation.User,
				Amount:   operation.Amount,
			}
			transferIn = append(transferIn, transfer)
		}
	}

	output = RetrieveUserInfoOutput{
		Balance:     balance,
		Inventory:   inventory,
		TransferIn:  transferIn,
		TransferOut: transferOut,
	}

	if err := s.Cache.Set(ctx, cacheKey, output); err != nil {
		s.Log.Error("Failed to cache user info",
			zap.String("op", op),
			zap.Error(err),
		)
	}

	s.Log.Info("Info successfully retrieved")

	return output, nil
}
