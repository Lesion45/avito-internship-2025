package pgdb

import (
	"context"
	"errors"
	"fmt"

	"avito-internship/internal/entity"
	"avito-internship/internal/repository/repoerrs"
	"avito-internship/pkg/postgres"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

var pgErr *pgconn.PgError

type UserRepository struct {
	*postgres.Postgres
}

func NewUserRepository(pg *postgres.Postgres) *UserRepository {
	return &UserRepository{pg}
}

func (r *UserRepository) AddUser(ctx context.Context, username string, password []byte) error {
	const op = "repository.UserRepository.CreateUser"

	query := `INSERT INTO users(username, password) VALUES(@username, @password)`
	args := pgx.NamedArgs{
		"username": username,
		"password": password,
	}

	_, err := r.Pool.Exec(ctx, query, args)
	if err != nil {
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return fmt.Errorf("%s: %w", op, repoerrs.ErrUserAlreadyExists)
		}
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (r *UserRepository) GetUserCredentials(ctx context.Context, username string) (entity.User, error) {
	const op = "repository.UserRepository.GetUserCredentials"

	query := `SELECT username, password FROM users WHERE username = @username`
	args := pgx.NamedArgs{
		"username": username,
	}

	var user entity.User

	err := r.Pool.QueryRow(ctx, query, args).Scan(&user.Username, &user.Password)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return entity.User{}, fmt.Errorf("%s: %w", op, repoerrs.ErrUserNotFound)
		}
		return entity.User{}, fmt.Errorf("%s: %w", op, err)
	}

	return user, nil
}

func (r *UserRepository) GetInfo(ctx context.Context, username string) (int, []entity.Operation, []entity.Inventory, error) {
	const op = "repository.UserRepository.GetInfo"

	tx, err := r.Pool.Begin(ctx)
	if err != nil {
		return 0, nil, nil, fmt.Errorf("%s: %w", op, err)
	}
	defer tx.Rollback(ctx)

	queryBalance := `SELECT balance FROM users WHERE username = @username`
	argsBalance := pgx.NamedArgs{"username": username}

	var balance int
	err = tx.QueryRow(ctx, queryBalance, argsBalance).Scan(&balance)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, nil, nil, fmt.Errorf("%s: %w", op, repoerrs.ErrUserNotFound)
		}
		return 0, nil, nil, fmt.Errorf("%s: %w", op, err)
	}

	queryOperations := `
		SELECT 
			u1.username AS user,
			u2.username AS counterparty,
			o.amount
		FROM operations o
		LEFT JOIN users u1 ON o.user_id = u1.id
		LEFT JOIN users u2 ON o.counterparty_id = u2.id
		WHERE (u1.username = @username OR u2.username = @username)
		AND o.type = 'transfer'`
	argsOperations := pgx.NamedArgs{"username": username}

	operations := []entity.Operation{}
	rows, err := tx.Query(ctx, queryOperations, argsOperations)
	if err != nil {
		return 0, nil, nil, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()

	for rows.Next() {
		var operation entity.Operation
		if err := rows.Scan(&operation.User, &operation.Counterparty, &operation.Amount); err != nil {
			return 0, nil, nil, fmt.Errorf("%s: %w", op, err)
		}

		operations = append(operations, operation)
	}
	if rows.Err() != nil {
		return 0, nil, nil, fmt.Errorf("%s: %w", op, rows.Err())
	}

	queryInventory := `
		SELECT 
			p.name AS product,
			i.quantity
		FROM inventory i
		JOIN products p ON i.product_id = p.id
		WHERE i.user_id = (SELECT id FROM users WHERE username = @username)`
	argsInventory := pgx.NamedArgs{"username": username}

	inventory := []entity.Inventory{}
	rows, err = tx.Query(ctx, queryInventory, argsInventory)
	if err != nil {
		return 0, nil, nil, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()

	for rows.Next() {
		var inv entity.Inventory
		if err := rows.Scan(&inv.Product, &inv.Quantity); err != nil {
			return 0, nil, nil, fmt.Errorf("%s: %w", op, err)
		}
		inventory = append(inventory, inv)
	}
	if rows.Err() != nil {
		return 0, nil, nil, fmt.Errorf("%s: %w", op, rows.Err())
	}

	if err := tx.Commit(ctx); err != nil {
		return 0, nil, nil, fmt.Errorf("%s: %w", op, err)
	}

	return balance, operations, inventory, nil
}
