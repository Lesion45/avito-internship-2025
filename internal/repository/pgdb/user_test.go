package pgdb

import (
	"context"
	"testing"

	"avito-internship/internal/entity"
	"avito-internship/pkg/postgres"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/assert"
)

func TestUserRepository_AddUser(t *testing.T) {
	type args struct {
		ctx      context.Context
		username string
		password []byte
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
				password: []byte("qwerty123"),
			},
			mockBehavior: func(m pgxmock.PgxPoolIface, args args) {
				m.ExpectExec("INSERT INTO users").
					WithArgs(args.username, args.password).
					WillReturnResult(pgxmock.NewResult("INSERT", 1))
			},
			wantErr: false,
		},
		{
			name: "User Already Exists",
			args: args{
				ctx:      context.Background(),
				username: "existing_user",
				password: []byte("qwerty123"),
			},
			mockBehavior: func(m pgxmock.PgxPoolIface, args args) {
				m.ExpectExec("INSERT INTO users").
					WithArgs(args.username, args.password).
					WillReturnError(&pgconn.PgError{Code: "23505"})
			},
			wantErr: true,
		},
		{
			name: "Query Execution Error",
			args: args{
				ctx:      context.Background(),
				username: "test_user",
				password: []byte("qwerty123"),
			},
			mockBehavior: func(m pgxmock.PgxPoolIface, args args) {
				m.ExpectExec("INSERT INTO users").
					WithArgs(args.username, args.password).
					WillReturnError(assert.AnError)
			},
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			poolMock, _ := pgxmock.NewPool()
			defer poolMock.Close()
			tc.mockBehavior(poolMock, tc.args)

			postgresMock := &postgres.Postgres{
				Pool: poolMock,
			}
			userRepoMock := NewUserRepository(postgresMock)

			err := userRepoMock.AddUser(tc.args.ctx, tc.args.username, tc.args.password)
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

func TestUserRepository_GetUserCredentials(t *testing.T) {
	type args struct {
		ctx      context.Context
		username string
	}

	type MockBehavior func(m pgxmock.PgxPoolIface, args args)

	testCases := []struct {
		name         string
		args         args
		mockBehavior MockBehavior
		wantUser     entity.User
		wantErr      bool
	}{
		{
			name: "OK",
			args: args{
				ctx:      context.Background(),
				username: "test_user",
			},
			mockBehavior: func(m pgxmock.PgxPoolIface, args args) {
				rows := pgxmock.NewRows([]string{"username", "password"}).
					AddRow("test_user", []byte("hashed_password"))
				m.ExpectQuery("SELECT username, password FROM users WHERE username = @username").
					WithArgs(args.username).
					WillReturnRows(rows)
			},
			wantUser: entity.User{
				Username: "test_user",
				Password: []byte("hashed_password"),
			},
			wantErr: false,
		},
		{
			name: "User Not Found",
			args: args{
				ctx:      context.Background(),
				username: "non_existent_user",
			},
			mockBehavior: func(m pgxmock.PgxPoolIface, args args) {
				m.ExpectQuery("SELECT username, password FROM users WHERE username = @username").
					WithArgs(args.username).
					WillReturnError(pgx.ErrNoRows)
			},
			wantUser: entity.User{},
			wantErr:  true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			poolMock, _ := pgxmock.NewPool()
			defer poolMock.Close()
			tc.mockBehavior(poolMock, tc.args)

			postgresMock := &postgres.Postgres{
				Pool: poolMock,
			}
			userRepoMock := NewUserRepository(postgresMock)

			user, err := userRepoMock.GetUserCredentials(tc.args.ctx, tc.args.username)

			if tc.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)

			assert.Equal(t, tc.wantUser, user)

			err = poolMock.ExpectationsWereMet()
			assert.NoError(t, err)
		})
	}
}

func TestUserRepository_GetInfo(t *testing.T) {
	type args struct {
		ctx      context.Context
		username string
	}

	type MockBehavior func(m pgxmock.PgxPoolIface, args args)

	testCases := []struct {
		name         string
		args         args
		mockBehavior MockBehavior
		wantBalance  int
		wantOps      []entity.Operation
		wantInv      []entity.Inventory
		wantErr      bool
	}{
		{
			name: "OK",
			args: args{
				ctx:      context.Background(),
				username: "test_user",
			},
			mockBehavior: func(m pgxmock.PgxPoolIface, args args) {
				m.ExpectBegin()
				m.ExpectQuery("SELECT balance FROM users").
					WithArgs(args.username).
					WillReturnRows(pgxmock.NewRows([]string{"balance"}).AddRow(100))

				m.ExpectQuery("SELECT.*FROM operations").
					WithArgs(args.username).
					WillReturnRows(pgxmock.NewRows([]string{"user", "counterparty", "amount"}).
						AddRow(args.username, "other_user", 50))

				m.ExpectQuery("SELECT.*FROM inventory").
					WithArgs(args.username).
					WillReturnRows(pgxmock.NewRows([]string{"product", "quantity"}).
						AddRow("item1", 10))

				m.ExpectCommit()
			},
			wantBalance: 100,
			wantOps: []entity.Operation{
				{User: "test_user", Counterparty: "other_user", Amount: 50},
			},
			wantInv: []entity.Inventory{
				{Product: "item1", Quantity: 10},
			},
			wantErr: false,
		},
		{
			name: "User Not Found",
			args: args{
				ctx:      context.Background(),
				username: "unknown_user",
			},
			mockBehavior: func(m pgxmock.PgxPoolIface, args args) {
				m.ExpectBegin()
				m.ExpectQuery("SELECT balance FROM users").
					WithArgs(args.username).
					WillReturnError(pgx.ErrNoRows)
				m.ExpectRollback()
			},
			wantErr: true,
		},
		{
			name: "Transaction Begin Error",
			args: args{
				ctx:      context.Background(),
				username: "test_user",
			},
			mockBehavior: func(m pgxmock.PgxPoolIface, args args) {
				m.ExpectBegin().WillReturnError(assert.AnError)
			},
			wantErr: true,
		},
		{
			name: "Operations Query Error",
			args: args{
				ctx:      context.Background(),
				username: "test_user",
			},
			mockBehavior: func(m pgxmock.PgxPoolIface, args args) {
				m.ExpectBegin()
				m.ExpectQuery("SELECT balance FROM users").
					WithArgs(args.username).
					WillReturnRows(pgxmock.NewRows([]string{"balance"}).AddRow(100))
				m.ExpectQuery("SELECT.*FROM operations").
					WithArgs(args.username).
					WillReturnError(assert.AnError)
				m.ExpectRollback()
			},
			wantErr: true,
		},
		{
			name: "User Without Operations and Inventory",
			args: args{
				ctx:      context.Background(),
				username: "empty_user",
			},
			mockBehavior: func(m pgxmock.PgxPoolIface, args args) {
				m.ExpectBegin()
				m.ExpectQuery("SELECT balance FROM users").
					WithArgs(args.username).
					WillReturnRows(pgxmock.NewRows([]string{"balance"}).AddRow(50))
				m.ExpectQuery("SELECT.*FROM operations").
					WithArgs(args.username).
					WillReturnRows(pgxmock.NewRows([]string{"user", "counterparty", "amount", "type"}))
				m.ExpectQuery("SELECT.*FROM inventory").
					WithArgs(args.username).
					WillReturnRows(pgxmock.NewRows([]string{"product", "quantity"}))
				m.ExpectCommit()
			},
			wantBalance: 50,
			wantOps:     []entity.Operation{},
			wantInv:     []entity.Inventory{},
			wantErr:     false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			poolMock, _ := pgxmock.NewPool()
			defer poolMock.Close()
			tc.mockBehavior(poolMock, tc.args)

			postgresMock := &postgres.Postgres{
				Pool: poolMock,
			}
			userRepoMock := NewUserRepository(postgresMock)

			balance, operations, inventory, err := userRepoMock.GetInfo(tc.args.ctx, tc.args.username)

			if tc.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tc.wantBalance, balance)
			assert.Equal(t, tc.wantOps, operations)
			assert.Equal(t, tc.wantInv, inventory)
		})
	}
}
