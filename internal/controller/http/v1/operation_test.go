package v1

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"avito-internship/internal/service"
	"avito-internship/internal/service/servicerrs"

	"github.com/gofiber/fiber/v2"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func Test_sendCoin(t *testing.T) {
	logger := zap.NewNop()
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockOperationService := service.NewMockOperation(ctrl)

	tests := []struct {
		name            string
		requestBody     map[string]interface{}
		mockServiceFunc func()
		expectedCode    int
		expectedBody    string
	}{
		{
			name:        "Successful transfer",
			requestBody: map[string]interface{}{"toUser": "recipient", "amount": 100},
			mockServiceFunc: func() {
				mockOperationService.EXPECT().
					TransferFunds(ctx, service.TransferFundsInput{
						Sender:    "sender",
						Recipient: "recipient",
						Amount:    100,
					}).
					Return(nil)
			},
			expectedCode: http.StatusOK,
			expectedBody: "",
		},
		{
			name:        "Recipient not found",
			requestBody: map[string]interface{}{"toUser": "recipient", "amount": 100},
			mockServiceFunc: func() {
				mockOperationService.EXPECT().
					TransferFunds(ctx, service.TransferFundsInput{
						Sender:    "sender",
						Recipient: "recipient",
						Amount:    100,
					}).
					Return(servicerrs.ErrRecipientNotFound)
			},
			expectedCode: http.StatusBadRequest,
			expectedBody: `{"errors":"recipient not found"}`,
		},
		{
			name:        "Insufficient funds",
			requestBody: map[string]interface{}{"toUser": "recipient", "amount": 100},
			mockServiceFunc: func() {
				mockOperationService.EXPECT().
					TransferFunds(ctx, service.TransferFundsInput{
						Sender:    "sender",
						Recipient: "recipient",
						Amount:    100,
					}).
					Return(servicerrs.ErrInsufficientFunds)
			},
			expectedCode: http.StatusBadRequest,
			expectedBody: `{"errors":"insufficient funds"}`,
		},
		{
			name:        "Internal server error",
			requestBody: map[string]interface{}{"toUser": "recipient", "amount": 100},
			mockServiceFunc: func() {
				mockOperationService.EXPECT().
					TransferFunds(ctx, service.TransferFundsInput{
						Sender:    "sender",
						Recipient: "recipient",
						Amount:    100,
					}).
					Return(errors.New("internal error"))
			},
			expectedCode: http.StatusInternalServerError,
			expectedBody: `{"errors":"internal error"}`,
		},
		{
			name:            "Invalid request (missing toUser)",
			requestBody:     map[string]interface{}{"amount": 100},
			mockServiceFunc: func() {},
			expectedCode:    http.StatusBadRequest,
			expectedBody:    `{"errors":"ToUser is a required"}`,
		},
		{
			name:            "Invalid request (missing amount)",
			requestBody:     map[string]interface{}{"toUser": "recipient"},
			mockServiceFunc: func() {},
			expectedCode:    http.StatusBadRequest,
			expectedBody:    `{"errors":"Amount is a required"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := fiber.New()
			r := operationRoutes{
				log:              logger,
				operationService: mockOperationService,
			}
			app.Post("/sendCoin", func(c *fiber.Ctx) error {
				c.Locals("username", "sender")
				return r.sendCoin(c, ctx)
			})

			tt.mockServiceFunc()

			reqBody, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest(http.MethodPost, "/sendCoin", bytes.NewReader(reqBody))
			req.Header.Set("Content-Type", "application/json")
			resp, _ := app.Test(req)

			assert.Equal(t, tt.expectedCode, resp.StatusCode)

			var bodyBytes bytes.Buffer
			_, _ = bodyBytes.ReadFrom(resp.Body)
			assert.Contains(t, bodyBytes.String(), tt.expectedBody)
		})
	}
}

func Test_buyProduct(t *testing.T) {
	logger := zap.NewNop()
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockOperationService := service.NewMockOperation(ctrl)

	tests := []struct {
		name            string
		item            string
		mockServiceFunc func()
		expectedCode    int
		expectedBody    string
	}{
		{
			name: "Successful purchase",
			item: "product1",
			mockServiceFunc: func() {
				mockOperationService.EXPECT().
					PurchaseProduct(ctx, service.PurchaseProductInput{
						Username: "user",
						Product:  "product1",
					}).
					Return(nil)
			},
			expectedCode: http.StatusOK,
			expectedBody: "",
		},
		{
			name: "Insufficient funds",
			item: "product1",
			mockServiceFunc: func() {
				mockOperationService.EXPECT().
					PurchaseProduct(ctx, service.PurchaseProductInput{
						Username: "user",
						Product:  "product1",
					}).
					Return(servicerrs.ErrInsufficientFunds)
			},
			expectedCode: http.StatusBadRequest,
			expectedBody: `{"errors":"insufficient funds"}`,
		},
		{
			name: "Product not found",
			item: "product1",
			mockServiceFunc: func() {
				mockOperationService.EXPECT().
					PurchaseProduct(ctx, service.PurchaseProductInput{
						Username: "user",
						Product:  "product1",
					}).
					Return(servicerrs.ErrProductNotFound)
			},
			expectedCode: http.StatusBadRequest,
			expectedBody: `{"errors":"product not found"}`,
		},
		{
			name: "Internal server error",
			item: "product1",
			mockServiceFunc: func() {
				mockOperationService.EXPECT().
					PurchaseProduct(ctx, service.PurchaseProductInput{
						Username: "user",
						Product:  "product1",
					}).
					Return(errors.New("internal error"))
			},
			expectedCode: http.StatusInternalServerError,
			expectedBody: `{"errors":"internal error"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := fiber.New()
			r := operationRoutes{
				log:              logger,
				operationService: mockOperationService,
			}
			app.Get("/buy/:item", func(c *fiber.Ctx) error {
				c.Locals("username", "user")
				return r.buyProduct(c, ctx)
			})

			tt.mockServiceFunc()

			req := httptest.NewRequest(http.MethodGet, "/buy/"+tt.item, nil)
			resp, _ := app.Test(req)

			assert.Equal(t, tt.expectedCode, resp.StatusCode)

			var bodyBytes bytes.Buffer
			_, _ = bodyBytes.ReadFrom(resp.Body)
			assert.Contains(t, bodyBytes.String(), tt.expectedBody)
		})
	}
}
