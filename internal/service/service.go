package service

import (
	"context"
	"time"

	"avito-internship/internal/cache"
	"avito-internship/internal/entity"
	"avito-internship/internal/repository"

	"go.uber.org/zap"
)

type Auth interface {
	Authorization(ctx context.Context, username, password string) (string, error)
	ValidateToken(token string) (string, error)
	handleUserNotFound(ctx context.Context, username, password, op string) (string, error)
	generateToken(username, op string) (string, error)
}

type UserCreateInput struct {
	Username string
	Password []byte
}

type RetrieveUserInfoInput struct {
	Username string
}

type RetrieveUserInfoOutput struct {
	Balance     int
	Inventory   []entity.Inventory
	TransferIn  []entity.Transfer
	TransferOut []entity.Transfer
}

type User interface {
	CreateUser(ctx context.Context, input UserCreateInput) error
	RetrieveUserInfo(ctx context.Context, input RetrieveUserInfoInput) (RetrieveUserInfoOutput, error)
}

type TransferFundsInput struct {
	Sender    string
	Recipient string
	Amount    int
}

type PurchaseProductInput struct {
	Username string
	Product  string
}

type Operation interface {
	TransferFunds(ctx context.Context, input TransferFundsInput) error
	PurchaseProduct(ctx context.Context, input PurchaseProductInput) error
}

type Services struct {
	Auth
	User
	Operation
}

type ServicesDependencies struct {
	Log      *zap.Logger
	Cache    cache.Cache
	Repos    *repository.Repositories
	TokenTTL time.Duration
	Salt     string
}

func NewServices(deps ServicesDependencies) *Services {
	return &Services{
		User:      NewUserService(deps.Log, deps.Cache, deps.Repos.User),
		Auth:      NewAuthService(deps.Log, deps.Repos.User, deps.TokenTTL, deps.Salt),
		Operation: NewOperationService(deps.Log, deps.Cache, deps.Repos.Operation),
	}
}
