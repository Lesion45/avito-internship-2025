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

func Test_Authorize(t *testing.T) {
	logger := zap.NewNop()
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAuthService := service.NewMockAuth(ctrl)

	tests := []struct {
		name         string
		requestBody  map[string]string
		mockAuthFunc func()
		expectedCode int
		expectedBody string
	}{
		{
			name:        "Successful authorization",
			requestBody: map[string]string{"username": "user", "password": "pass"},
			mockAuthFunc: func() {
				mockAuthService.EXPECT().Authorization(ctx, "user", "pass").Return("valid-token", nil)
			},
			expectedCode: http.StatusOK,
			expectedBody: `{"token":"valid-token"}`,
		},
		{
			name:        "Invalid credentials",
			requestBody: map[string]string{"username": "user", "password": "wrong"},
			mockAuthFunc: func() {
				mockAuthService.EXPECT().Authorization(ctx, "user", "wrong").Return("", servicerrs.ErrInvalidCredentials)
			},
			expectedCode: http.StatusBadRequest,
			expectedBody: `{"errors":"invalid credentials"}`,
		},
		{
			name:        "Internal server error",
			requestBody: map[string]string{"username": "user", "password": "pass"},
			mockAuthFunc: func() {
				mockAuthService.EXPECT().Authorization(ctx, "user", "pass").Return("", errors.New("unexpected error"))
			},
			expectedCode: http.StatusInternalServerError,
			expectedBody: `{"errors":"internal error"}`,
		},
		{
			name:         "Empty request",
			requestBody:  map[string]string{},
			mockAuthFunc: func() {},
			expectedCode: http.StatusBadRequest,
			expectedBody: `{"errors":"Username is a required, Password is a required"}`,
		},
		{
			name:         "Missing password field",
			requestBody:  map[string]string{"username": "user"},
			mockAuthFunc: func() {},
			expectedCode: http.StatusBadRequest,
			expectedBody: `{"errors":"Password is a required"}`,
		},
		{
			name:         "Missing username field",
			requestBody:  map[string]string{"password": "pass"},
			mockAuthFunc: func() {},
			expectedCode: http.StatusBadRequest,
			expectedBody: `{"errors":"Username is a required"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := fiber.New()
			r := &AuthRoutes{log: logger, authService: mockAuthService}
			app.Post("/auth", func(c *fiber.Ctx) error { return r.authorize(c, ctx) })

			tt.mockAuthFunc()

			reqBody, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest(http.MethodPost, "/auth", bytes.NewReader(reqBody))
			req.Header.Set("Content-Type", "application/json")
			resp, _ := app.Test(req)

			assert.Equal(t, tt.expectedCode, resp.StatusCode)

			var bodyBytes bytes.Buffer
			_, _ = bodyBytes.ReadFrom(resp.Body)
			assert.Contains(t, bodyBytes.String(), tt.expectedBody)
		})
	}
}
