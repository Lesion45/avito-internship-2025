package pgdb

import (
	"context"
	"errors"
	"fmt"

	"avito-internship/internal/model"
	"avito-internship/internal/repository/repoerrs"
	"avito-internship/pkg/postgres"

	"github.com/jackc/pgx/v5"
)

type OperationRepository struct {
	*postgres.Postgres
}

func NewOperationRepository(pg *postgres.Postgres) *OperationRepository {
	return &OperationRepository{pg}
}

func (r *OperationRepository) SaveTransfer(ctx context.Context, sender string, recipient string, amount int) error {
	const op = "repository.OperationRepository.Transfer"

	tx, err := r.Pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	defer tx.Rollback(ctx)

	var senderID int
	var senderBalance int
	getSenderQuery := `SELECT id, balance FROM users WHERE username = @username`
	getSenderArgs := pgx.NamedArgs{
		"username": sender,
	}

	err = tx.QueryRow(ctx, getSenderQuery, getSenderArgs).Scan(&senderID, &senderBalance)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("%s: %w", op, repoerrs.ErrUserNotFound)
		}
		return fmt.Errorf("%s: %w", op, err)
	}

	if senderBalance < amount {
		return fmt.Errorf("%s: %w", op, repoerrs.ErrInsufficientFunds)
	}

	var recipientID int
	getRecipientIDQuery := `SELECT id FROM users WHERE username = @username`
	getRecipientIDArgs := pgx.NamedArgs{
		"username": recipient,
	}

	err = tx.QueryRow(ctx, getRecipientIDQuery, getRecipientIDArgs).Scan(&recipientID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("%s: %w", op, repoerrs.ErrUserNotFound)
		}
		return fmt.Errorf("%s: %w", op, err)
	}

	updateSenderQuery := `UPDATE users SET balance = balance - @amount WHERE id = @id`
	updateSenderArgs := pgx.NamedArgs{
		"id":     senderID,
		"amount": amount,
	}

	_, err = tx.Exec(ctx, updateSenderQuery, updateSenderArgs)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	updateRecipientQuery := `UPDATE users SET balance = balance + @amount WHERE id = @id`
	updateRecipientArgs := pgx.NamedArgs{
		"id":     recipientID,
		"amount": amount,
	}

	_, err = tx.Exec(ctx, updateRecipientQuery, updateRecipientArgs)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	operationQuery := `
        INSERT INTO operations (user_id, amount, type, counterparty_id)
        VALUES (@user_id, @amount, @type, @counterparty_id)
    `
	operationArgs := pgx.NamedArgs{
		"user_id":         senderID,
		"amount":          amount,
		"type":            model.OperationTypeTransfer,
		"counterparty_id": recipientID,
	}

	_, err = tx.Exec(ctx, operationQuery, operationArgs)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (r *OperationRepository) SavePurchase(ctx context.Context, username string, product string) error {
	const op = "repository.OperationRepository.Purchase"

	tx, err := r.Pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	defer tx.Rollback(ctx)

	var userID int
	var userBalance int
	query := `SELECT id, balance FROM users WHERE username = @username`
	args := pgx.NamedArgs{
		"username": username,
	}

	err = tx.QueryRow(ctx, query, args).Scan(&userID, &userBalance)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("%s: %w", op, repoerrs.ErrUserNotFound)
		}
		return fmt.Errorf("%s: %w", op, err)
	}

	var productID int
	var productPrice int
	query = `SELECT id, price FROM products WHERE name = @product`
	args = pgx.NamedArgs{
		"product": product,
	}

	err = tx.QueryRow(ctx, query, args).Scan(&productID, &productPrice)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("%s: %w", op, repoerrs.ErrProductNotFound)
		}
		return fmt.Errorf("%s: %w", op, err)
	}

	if userBalance < productPrice {
		return fmt.Errorf("%s: %w", op, repoerrs.ErrInsufficientFunds)
	}

	updateBalanceQuery := `UPDATE users SET balance = balance - @price WHERE id = @id`
	updateBalanceArgs := pgx.NamedArgs{
		"id":    userID,
		"price": productPrice,
	}

	_, err = tx.Exec(ctx, updateBalanceQuery, updateBalanceArgs)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	upsertInventoryQuery := `
        INSERT INTO inventory (user_id, product_id, quantity)
        VALUES (@user_id, @product_id, 1)
        ON CONFLICT (user_id, product_id) DO UPDATE
        SET quantity = inventory.quantity + 1
    `
	upsertInventoryArgs := pgx.NamedArgs{
		"user_id":    userID,
		"product_id": productID,
	}

	_, err = tx.Exec(ctx, upsertInventoryQuery, upsertInventoryArgs)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	operationQuery := `
        INSERT INTO operations (user_id, amount, type, product_id)
        VALUES (@user_id, @amount, @type, @product_id)
    `
	operationArgs := pgx.NamedArgs{
		"user_id":    userID,
		"amount":     productPrice,
		"type":       "purchase",
		"product_id": productID,
	}

	fmt.Println(operationArgs)

	_, err = tx.Exec(ctx, operationQuery, operationArgs)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}
