package repository

import (
	"context"

	"avito-internship/internal/entity"
	"avito-internship/internal/repository/pgdb"
	"avito-internship/pkg/postgres"
)

type User interface {
	AddUser(ctx context.Context, username string, password []byte) error
	GetUserCredentials(ctx context.Context, username string) (entity.User, error)
	GetInfo(ctx context.Context, username string) (int, []entity.Operation, []entity.Inventory, error)
}

type Operation interface {
	SaveTransfer(ctx context.Context, sender string, recipient string, amount int) error
	SavePurchase(ctx context.Context, username string, product string) error
}

type Repositories struct {
	User
	Operation
}

func NewRepositories(pg *postgres.Postgres) *Repositories {
	return &Repositories{
		User:      pgdb.NewUserRepository(pg),
		Operation: pgdb.NewOperationRepository(pg),
	}
}
