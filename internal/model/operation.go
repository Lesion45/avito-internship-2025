package model

import (
	"time"

	"github.com/google/uuid"
)

const (
	OperationTypeTransfer = "transfer"
	OperationTypePurchase = "purchase"
)

type Operation struct {
	ID             uuid.UUID `db:"id"`
	UserID         int       `db:"user_id"`
	Amount         int       `db:"amount"`
	Type           string    `db:"type"`
	CounterpartyID int       `db:"counterparty_id"`
	ProductID      int       `db:"product_id"`
	CreatedAt      time.Time `db:"created_at"`
}
