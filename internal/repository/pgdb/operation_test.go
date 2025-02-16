package pgdb

import (
	"context"
	"errors"
	"testing"

	"avito-internship/internal/model"
	"avito-internship/pkg/postgres"

	"github.com/jackc/pgx/v5"
	"github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/assert"
)

func TestUserRepository_SaveTransfer(t *testing.T) {
	type args struct {
		ctx       context.Context
		sender    string
		recipient string
		amount    int
	}

	type MockBehavior func(m pgxmock.PgxPoolIface, args args)

	testCases := []struct {
		name         string
		args         args
		mockBehavior MockBehavior
		wantErr      bool
	}{
		{
			name: "OK",
			args: args{
				ctx:       context.Background(),
				sender:    "sender_user",
				recipient: "recipient_user",
				amount:    100,
			},
			mockBehavior: func(m pgxmock.PgxPoolIface, args args) {
				m.ExpectBegin()

				rows := pgxmock.NewRows([]string{"id", "balance"}).
					AddRow(1, 500)
				m.ExpectQuery("SELECT id, balance FROM users WHERE username = @username").
					WithArgs(args.sender).
					WillReturnRows(rows)

				recipientRows := pgxmock.NewRows([]string{"id"}).
					AddRow(2)
				m.ExpectQuery("SELECT id FROM users WHERE username = @username").
					WithArgs(args.recipient).
					WillReturnRows(recipientRows)

				m.ExpectExec("UPDATE users SET balance = balance - @amount WHERE id = @id").
					WithArgs(args.amount, 1).
					WillReturnResult(pgxmock.NewResult("UPDATE", 1))

				m.ExpectExec("UPDATE users SET balance = balance \\+ @amount WHERE id = @id").
					WithArgs(args.amount, 2).
					WillReturnResult(pgxmock.NewResult("UPDATE", 1))

				m.ExpectExec("INSERT INTO operations \\(user_id, amount, type, counterparty_id\\) VALUES \\(@user_id, @amount, @type, @counterparty_id\\)").
					WithArgs(1, args.amount, model.OperationTypeTransfer, 2).
					WillReturnResult(pgxmock.NewResult("INSERT", 1))

				m.ExpectCommit()
			},
			wantErr: false,
		},
		{
			name: "Sender Not Found",
			args: args{
				ctx:       context.Background(),
				sender:    "non_existent_sender",
				recipient: "recipient_user",
				amount:    100,
			},
			mockBehavior: func(m pgxmock.PgxPoolIface, args args) {
				m.ExpectBegin()

				m.ExpectQuery("SELECT id, balance FROM users WHERE username = @username").
					WithArgs(args.sender).
					WillReturnError(pgx.ErrNoRows)

				m.ExpectRollback()
			},
			wantErr: true,
		},
		{
			name: "Recipient Not Found",
			args: args{
				ctx:       context.Background(),
				sender:    "sender_user",
				recipient: "non_existent_recipient",
				amount:    100,
			},
			mockBehavior: func(m pgxmock.PgxPoolIface, args args) {
				m.ExpectBegin()

				senderRows := pgxmock.NewRows([]string{"id", "balance"}).
					AddRow(1, 1000)
				m.ExpectQuery("SELECT id, balance FROM users WHERE username = @username").
					WithArgs(args.sender).
					WillReturnRows(senderRows)

				m.ExpectQuery("SELECT id, balance FROM users WHERE username = @username").
					WithArgs(args.recipient).
					WillReturnError(pgx.ErrNoRows)

				m.ExpectRollback()
			},
			wantErr: true,
		},
		{
			name: "Insufficient Funds",
			args: args{
				ctx:       context.Background(),
				sender:    "sender_user",
				recipient: "recipient_user",
				amount:    1500,
			},
			mockBehavior: func(m pgxmock.PgxPoolIface, args args) {
				m.ExpectBegin()

				senderRows := pgxmock.NewRows([]string{"id", "balance"}).
					AddRow(1, 1000)
				m.ExpectQuery("SELECT id, balance FROM users WHERE username = @username").
					WithArgs(args.sender).
					WillReturnRows(senderRows)

				m.ExpectRollback()
			},
			wantErr: true,
		},
		{
			name: "Update Sender Balance Error",
			args: args{
				ctx:       context.Background(),
				sender:    "sender_user",
				recipient: "recipient_user",
				amount:    100,
			},
			mockBehavior: func(m pgxmock.PgxPoolIface, args args) {
				m.ExpectBegin()

				senderRows := pgxmock.NewRows([]string{"id", "balance"}).
					AddRow(1, 1000)
				m.ExpectQuery("SELECT id, balance FROM users WHERE username = @username").
					WithArgs(args.sender).
					WillReturnRows(senderRows)

				recipientRows := pgxmock.NewRows([]string{"id", "balance"}).
					AddRow(2, 500)
				m.ExpectQuery("SELECT id, balance FROM users WHERE username = @username").
					WithArgs(args.recipient).
					WillReturnRows(recipientRows)

				m.ExpectExec("UPDATE users SET balance = balance - @amount WHERE id = @id").
					WithArgs(pgx.NamedArgs{"id": 1, "amount": args.amount}).
					WillReturnError(errors.New("update sender balance error"))

				m.ExpectRollback()
			},
			wantErr: true,
		},
		{
			name: "Update Recipient Balance Error",
			args: args{
				ctx:       context.Background(),
				sender:    "sender_user",
				recipient: "recipient_user",
				amount:    100,
			},
			mockBehavior: func(m pgxmock.PgxPoolIface, args args) {
				m.ExpectBegin()

				senderRows := pgxmock.NewRows([]string{"id", "balance"}).
					AddRow(1, 1000)
				m.ExpectQuery("SELECT id, balance FROM users WHERE username = @username").
					WithArgs(args.sender).
					WillReturnRows(senderRows)

				recipientRows := pgxmock.NewRows([]string{"id", "balance"}).
					AddRow(2, 500)
				m.ExpectQuery("SELECT id, balance FROM users WHERE username = @username").
					WithArgs(args.recipient).
					WillReturnRows(recipientRows)

				m.ExpectExec("UPDATE users SET balance = balance - @amount WHERE id = @id").
					WithArgs(pgx.NamedArgs{"id": 1, "amount": args.amount}).
					WillReturnResult(pgxmock.NewResult("UPDATE", 1))

				m.ExpectExec("UPDATE users SET balance = balance + @amount WHERE id = @id").
					WithArgs(pgx.NamedArgs{"id": 2, "amount": args.amount}).
					WillReturnError(errors.New("update recipient balance error"))

				m.ExpectRollback()
			},
			wantErr: true,
		},
		{
			name: "Insert Operation Error",
			args: args{
				ctx:       context.Background(),
				sender:    "sender_user",
				recipient: "recipient_user",
				amount:    100,
			},
			mockBehavior: func(m pgxmock.PgxPoolIface, args args) {
				m.ExpectBegin()

				senderRows := pgxmock.NewRows([]string{"id", "balance"}).
					AddRow(1, 1000)
				m.ExpectQuery("SELECT id, balance FROM users WHERE username = @username").
					WithArgs(args.sender).
					WillReturnRows(senderRows)

				recipientRows := pgxmock.NewRows([]string{"id", "balance"}).
					AddRow(2, 500)
				m.ExpectQuery("SELECT id, balance FROM users WHERE username = @username").
					WithArgs(args.recipient).
					WillReturnRows(recipientRows)

				m.ExpectExec("UPDATE users SET balance = balance - @amount WHERE id = @id").
					WithArgs(pgx.NamedArgs{"id": 1, "amount": args.amount}).
					WillReturnResult(pgxmock.NewResult("UPDATE", 1))

				m.ExpectExec("UPDATE users SET balance = balance + @amount WHERE id = @id").
					WithArgs(pgx.NamedArgs{"id": 2, "amount": args.amount}).
					WillReturnResult(pgxmock.NewResult("UPDATE", 1))

				m.ExpectExec("INSERT INTO operations").
					WithArgs(pgx.NamedArgs{
						"user_id":         1,
						"amount":          args.amount,
						"type":            "transfer",
						"counterparty_id": 2,
					}).
					WillReturnError(errors.New("insert operation error"))

				m.ExpectRollback()
			},
			wantErr: true,
		},
		{
			name: "Commit Transaction Error",
			args: args{
				ctx:       context.Background(),
				sender:    "sender_user",
				recipient: "recipient_user",
				amount:    100,
			},
			mockBehavior: func(m pgxmock.PgxPoolIface, args args) {
				m.ExpectBegin()

				senderRows := pgxmock.NewRows([]string{"id", "balance"}).
					AddRow(1, 1000)
				m.ExpectQuery("SELECT id, balance FROM wusers WHERE username = @username").
					WithArgs(args.sender).
					WillReturnRows(senderRows)

				recipientRows := pgxmock.NewRows([]string{"id", "balance"}).
					AddRow(2, 500)
				m.ExpectQuery("SELECT id, balance FROM users WHERE username = @username").
					WithArgs(args.recipient).
					WillReturnRows(recipientRows)

				m.ExpectExec("UPDATE users SET balance = balance - @amount WHERE id = @id").
					WithArgs(pgx.NamedArgs{"id": 1, "amount": args.amount}).
					WillReturnResult(pgxmock.NewResult("UPDATE", 1))

				m.ExpectExec("UPDATE users SET balance = balance + @amount WHERE id = @id").
					WithArgs(pgx.NamedArgs{"id": 2, "amount": args.amount}).
					WillReturnResult(pgxmock.NewResult("UPDATE", 1))

				m.ExpectExec("INSERT INTO operations").
					WithArgs(pgx.NamedArgs{
						"user_id":         1,
						"amount":          args.amount,
						"type":            "transfer",
						"counterparty_id": 2,
					}).
					WillReturnResult(pgxmock.NewResult("INSERT", 1))

				m.ExpectCommit().WillReturnError(errors.New("commit error"))
			},
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			poolMock, err := pgxmock.NewPool()
			if err != nil {
				t.Fatalf("Failed to create mock pool: %v", err)
			}
			defer poolMock.Close()

			tc.mockBehavior(poolMock, tc.args)

			postgresMock := &postgres.Postgres{
				Pool: poolMock,
			}
			operationRepoMock := NewOperationRepository(postgresMock)

			err = operationRepoMock.SaveTransfer(tc.args.ctx, tc.args.sender, tc.args.recipient, tc.args.amount)

			if tc.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)

			err = poolMock.ExpectationsWereMet()
			assert.NoError(t, err)
		})
	}
}

func TestOperationRepository_SavePurchase(t *testing.T) {
	type args struct {
		ctx      context.Context
		username string
		product  string
	}

	type MockBehavior func(m pgxmock.PgxPoolIface, args args)

	testCases := []struct {
		name         string
		args         args
		mockBehavior MockBehavior
		wantErr      bool
	}{
		{
			name: "OK",
			args: args{
				ctx:      context.Background(),
				username: "test_user",
				product:  "test_product",
			},
			mockBehavior: func(m pgxmock.PgxPoolIface, args args) {
				m.ExpectBegin()

				userRows := pgxmock.NewRows([]string{"id", "balance"}).
					AddRow(1, 1000)
				m.ExpectQuery("SELECT id, balance FROM users WHERE username = @username").
					WithArgs(pgx.NamedArgs{"username": args.username}).
					WillReturnRows(userRows)

				productRows := pgxmock.NewRows([]string{"id", "price"}).
					AddRow(1, 100)
				m.ExpectQuery("SELECT id, price FROM products WHERE name = @product").
					WithArgs(pgx.NamedArgs{"product": args.product}).
					WillReturnRows(productRows)

				m.ExpectExec("UPDATE users SET balance = balance - @price WHERE id = @id").
					WithArgs(pgx.NamedArgs{"id": 1, "price": 100}).
					WillReturnResult(pgxmock.NewResult("UPDATE", 1))

				m.ExpectExec("INSERT INTO inventory").
					WithArgs(pgx.NamedArgs{"user_id": 1, "product_id": 1}).
					WillReturnResult(pgxmock.NewResult("INSERT", 1))

				m.ExpectExec("INSERT INTO operations").
					WithArgs(pgx.NamedArgs{
						"user_id":    1,
						"amount":     100,
						"type":       model.OperationTypePurchase,
						"product_id": 1,
					}).
					WillReturnResult(pgxmock.NewResult("INSERT", 1))

				m.ExpectCommit()
			},
			wantErr: false,
		},
		{
			name: "User Not Found",
			args: args{
				ctx:      context.Background(),
				username: "non_existent_user",
				product:  "test_product",
			},
			mockBehavior: func(m pgxmock.PgxPoolIface, args args) {
				m.ExpectBegin()

				m.ExpectQuery("SELECT id, balance FROM users WHERE username = @username").
					WithArgs(pgx.NamedArgs{"username": args.username}).
					WillReturnError(pgx.ErrNoRows)

				m.ExpectRollback()
			},
			wantErr: true,
		},
		{
			name: "Product Not Found",
			args: args{
				ctx:      context.Background(),
				username: "test_user",
				product:  "non_existent_product",
			},
			mockBehavior: func(m pgxmock.PgxPoolIface, args args) {
				m.ExpectBegin()

				userRows := pgxmock.NewRows([]string{"id", "balance"}).
					AddRow(1, 1000)
				m.ExpectQuery("SELECT id, balance FROM users WHERE username = @username").
					WithArgs(pgx.NamedArgs{"username": args.username}).
					WillReturnRows(userRows)

				m.ExpectQuery("SELECT id, price FROM products WHERE name = @product").
					WithArgs(pgx.NamedArgs{"product": args.product}).
					WillReturnError(pgx.ErrNoRows)

				m.ExpectRollback()
			},
			wantErr: true,
		},
		{
			name: "Insufficient Funds",
			args: args{
				ctx:      context.Background(),
				username: "test_user",
				product:  "test_product",
			},
			mockBehavior: func(m pgxmock.PgxPoolIface, args args) {
				m.ExpectBegin()

				userRows := pgxmock.NewRows([]string{"id", "balance"}).
					AddRow(1, 50) // Недостаточно средств
				m.ExpectQuery("SELECT id, balance FROM users WHERE username = @username").
					WithArgs(pgx.NamedArgs{"username": args.username}).
					WillReturnRows(userRows)

				productRows := pgxmock.NewRows([]string{"id", "price"}).
					AddRow(1, 100)
				m.ExpectQuery("SELECT id, price FROM products WHERE name = @product").
					WithArgs(pgx.NamedArgs{"product": args.product}).
					WillReturnRows(productRows)

				m.ExpectRollback()
			},
			wantErr: true,
		},
		{
			name: "Update Balance Error",
			args: args{
				ctx:      context.Background(),
				username: "test_user",
				product:  "test_product",
			},
			mockBehavior: func(m pgxmock.PgxPoolIface, args args) {
				m.ExpectBegin()

				userRows := pgxmock.NewRows([]string{"id", "balance"}).
					AddRow(1, 1000)
				m.ExpectQuery("SELECT id, balance FROM users WHERE username = @username").
					WithArgs(pgx.NamedArgs{"username": args.username}).
					WillReturnRows(userRows)

				productRows := pgxmock.NewRows([]string{"id", "price"}).
					AddRow(1, 100)
				m.ExpectQuery("SELECT id, price FROM products WHERE name = @product").
					WithArgs(pgx.NamedArgs{"product": args.product}).
					WillReturnRows(productRows)

				m.ExpectExec("UPDATE users SET balance = balance - @price WHERE id = @id").
					WithArgs(pgx.NamedArgs{"id": 1, "price": 100}).
					WillReturnError(errors.New("update balance error"))

				m.ExpectRollback()
			},
			wantErr: true,
		},
		{
			name: "Insert Inventory Error",
			args: args{
				ctx:      context.Background(),
				username: "test_user",
				product:  "test_product",
			},
			mockBehavior: func(m pgxmock.PgxPoolIface, args args) {
				m.ExpectBegin()

				userRows := pgxmock.NewRows([]string{"id", "balance"}).
					AddRow(1, 1000)
				m.ExpectQuery("SELECT id, balance FROM users WHERE username = @username").
					WithArgs(pgx.NamedArgs{"username": args.username}).
					WillReturnRows(userRows)

				productRows := pgxmock.NewRows([]string{"id", "price"}).
					AddRow(1, 100)
				m.ExpectQuery("SELECT id, price FROM products WHERE name = @product").
					WithArgs(pgx.NamedArgs{"product": args.product}).
					WillReturnRows(productRows)

				m.ExpectExec("UPDATE users SET balance = balance - @price WHERE id = @id").
					WithArgs(pgx.NamedArgs{"id": 1, "price": 100}).
					WillReturnResult(pgxmock.NewResult("UPDATE", 1))

				m.ExpectExec("INSERT INTO inventory").
					WithArgs(pgx.NamedArgs{"user_id": 1, "product_id": 1}).
					WillReturnError(errors.New("insert inventory error"))

				m.ExpectRollback()
			},
			wantErr: true,
		},
		{
			name: "Insert Operation Error",
			args: args{
				ctx:      context.Background(),
				username: "test_user",
				product:  "test_product",
			},
			mockBehavior: func(m pgxmock.PgxPoolIface, args args) {
				m.ExpectBegin()

				userRows := pgxmock.NewRows([]string{"id", "balance"}).
					AddRow(1, 1000)
				m.ExpectQuery("SELECT id, balance FROM users WHERE username = @username").
					WithArgs(pgx.NamedArgs{"username": args.username}).
					WillReturnRows(userRows)

				productRows := pgxmock.NewRows([]string{"id", "price"}).
					AddRow(1, 100)
				m.ExpectQuery("SELECT id, price FROM products WHERE name = @product").
					WithArgs(pgx.NamedArgs{"product": args.product}).
					WillReturnRows(productRows)

				m.ExpectExec("UPDATE users SET balance = balance - @price WHERE id = @id").
					WithArgs(pgx.NamedArgs{"id": 1, "price": 100}).
					WillReturnResult(pgxmock.NewResult("UPDATE", 1))

				m.ExpectExec("INSERT INTO inventory").
					WithArgs(pgx.NamedArgs{"user_id": 1, "product_id": 1}).
					WillReturnResult(pgxmock.NewResult("INSERT", 1))

				m.ExpectExec("INSERT INTO operations").
					WithArgs(pgx.NamedArgs{
						"user_id":    1,
						"amount":     100,
						"type":       model.OperationTypePurchase,
						"product_id": 1,
					}).
					WillReturnError(errors.New("insert operation error"))

				m.ExpectRollback()
			},
			wantErr: true,
		},
		{
			name: "Commit Transaction Error",
			args: args{
				ctx:      context.Background(),
				username: "test_user",
				product:  "test_product",
			},
			mockBehavior: func(m pgxmock.PgxPoolIface, args args) {
				m.ExpectBegin()

				userRows := pgxmock.NewRows([]string{"id", "balance"}).
					AddRow(1, 1000)
				m.ExpectQuery("SELECT id, balance FROM users WHERE username = @username").
					WithArgs(pgx.NamedArgs{"username": args.username}).
					WillReturnRows(userRows)

				productRows := pgxmock.NewRows([]string{"id", "price"}).
					AddRow(1, 100)
				m.ExpectQuery("SELECT id, price FROM products WHERE name = @product").
					WithArgs(pgx.NamedArgs{"product": args.product}).
					WillReturnRows(productRows)

				m.ExpectExec("UPDATE users SET balance = balance - @price WHERE id = @id").
					WithArgs(pgx.NamedArgs{"id": 1, "price": 100}).
					WillReturnResult(pgxmock.NewResult("UPDATE", 1))

				m.ExpectExec("INSERT INTO inventory").
					WithArgs(pgx.NamedArgs{"user_id": 1, "product_id": 1}).
					WillReturnResult(pgxmock.NewResult("INSERT", 1))

				m.ExpectExec("INSERT INTO operations").
					WithArgs(pgx.NamedArgs{
						"user_id":    1,
						"amount":     100,
						"type":       model.OperationTypePurchase,
						"product_id": 1,
					}).
					WillReturnResult(pgxmock.NewResult("INSERT", 1))

				m.ExpectCommit().WillReturnError(errors.New("commit error"))
			},
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			poolMock, err := pgxmock.NewPool()
			if err != nil {
				t.Fatalf("Failed to create mock pool: %v", err)
			}
			defer poolMock.Close()

			tc.mockBehavior(poolMock, tc.args)

			postgresMock := &postgres.Postgres{
				Pool: poolMock,
			}
			operationRepo := NewOperationRepository(postgresMock)

			err = operationRepo.SavePurchase(tc.args.ctx, tc.args.username, tc.args.product)

			if tc.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)

			err = poolMock.ExpectationsWereMet()
			assert.NoError(t, err)
		})
	}
}
