package service

import (
	"context"
	"errors"
	"fmt"

	"avito-internship/internal/cache"
	"avito-internship/internal/repository"
	"avito-internship/internal/repository/repoerrs"
	"avito-internship/internal/service/servicerrs"

	"go.uber.org/zap"
)

type OperationService struct {
	log   *zap.Logger
	cache cache.Cache
	repo  repository.Operation
}

func NewOperationService(log *zap.Logger, cache cache.Cache, repo repository.Operation) *OperationService {
	return &OperationService{
		log:   log,
		cache: cache,
		repo:  repo,
	}
}

func (s *OperationService) TransferFunds(ctx context.Context, input TransferFundsInput) error {
	const op = "service.OperationService.TransferFunds"

	s.log.Info("attempting to transfer funds")

	err := s.repo.SaveTransfer(ctx, input.Sender, input.Recipient, input.Amount)
	if err != nil {
		if errors.Is(err, repoerrs.ErrUserNotFound) {
			s.log.Error("recipient not found",
				zap.String("op", op),
				zap.String("recipient", input.Recipient),
			)

			return fmt.Errorf("%s: %w", op, servicerrs.ErrRecipientNotFound)
		} else if errors.Is(err, repoerrs.ErrInsufficientFunds) {
			s.log.Warn("insufficient funds",
				zap.String("op", op),
				zap.String("sender", input.Sender),
			)

			return fmt.Errorf("%s: %w", op, servicerrs.ErrInsufficientFunds)
		}

		s.log.Error("failed to save transfer to database",
			zap.String("op", op),
			zap.Error(err),
		)

		return fmt.Errorf("%s: %w", op, err)
	}

	senderCacheKey := fmt.Sprintf("user_info:%s", input.Sender)
	recipientCacheKey := fmt.Sprintf("user_info:%s", input.Recipient)

	if err := s.cache.Del(ctx, senderCacheKey, recipientCacheKey); err != nil {
		s.log.Error("failed to invalidate cache",
			zap.String("op", op),
			zap.Error(err),
		)
	}

	s.log.Info("transfer successfully completed")

	return nil
}

func (s *OperationService) PurchaseProduct(ctx context.Context, input PurchaseProductInput) error {
	const op = "service.OperationService.PurchaseProduct"

	s.log.Info("attempting to purchase product")

	err := s.repo.SavePurchase(ctx, input.Username, input.Product)
	if err != nil {
		if errors.Is(err, repoerrs.ErrUserNotFound) {
			s.log.Error("Customer not found",
				zap.String("op", op),
				zap.String("customer", input.Username),
			)

			return fmt.Errorf("%s: %w", op, servicerrs.ErrCustomerNotFound)
		} else if errors.Is(err, repoerrs.ErrProductNotFound) {
			s.log.Warn("product not found",
				zap.String("op", op),
				zap.String("product", input.Product),
			)

			return fmt.Errorf("%s: %w", op, servicerrs.ErrProductNotFound)
		} else if errors.Is(err, repoerrs.ErrInsufficientFunds) {
			s.log.Warn("insufficient funds",
				zap.String("op", op),
				zap.String("customer", input.Username),
			)

			return fmt.Errorf("%s: %w", op, servicerrs.ErrInsufficientFunds)
		}
	}

	customerCacheKey := fmt.Sprintf("user_info:%s", input.Username)

	if err := s.cache.Del(ctx, customerCacheKey); err != nil {
		s.log.Error("failed to invalidate cache",
			zap.String("op", op),
			zap.Error(err),
		)
	}

	s.log.Info("Purchase successfully completed")

	return nil
}
