package v1

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"avito-internship/internal/entity"
	"avito-internship/internal/service"
	"avito-internship/internal/service/servicerrs"

	"github.com/gofiber/fiber/v2"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestGetInfo(t *testing.T) {
	logger := zap.NewNop()
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserService := service.NewMockUser(ctrl)

	tests := []struct {
		name         string
		username     string
		mockUserFunc func()
		expectedCode int
		expectedBody string
	}{
		{
			name:     "Successful user info retrieval",
			username: "user1",
			mockUserFunc: func() {
				mockUserService.EXPECT().RetrieveUserInfo(ctx, service.RetrieveUserInfoInput{Username: "user1"}).Return(service.RetrieveUserInfoOutput{
					Balance:     100,
					Inventory:   []entity.Inventory{{Product: "item1", Quantity: 2}},
					TransferIn:  []entity.Transfer{{Username: "user2", Amount: 50}},
					TransferOut: []entity.Transfer{{Username: "user3", Amount: 30}},
				}, nil)
			},
			expectedCode: http.StatusOK,
			expectedBody: `{"coins":100`,
		},
		{
			name:     "User not found",
			username: "unknown",
			mockUserFunc: func() {
				mockUserService.EXPECT().RetrieveUserInfo(ctx, service.RetrieveUserInfoInput{Username: "unknown"}).Return(service.RetrieveUserInfoOutput{}, servicerrs.ErrUserNotFound)
			},
			expectedCode: http.StatusBadRequest,
			expectedBody: `{"errors":"user not found"}`,
		},
		{
			name:     "Internal server error",
			username: "user1",
			mockUserFunc: func() {
				mockUserService.EXPECT().RetrieveUserInfo(ctx, service.RetrieveUserInfoInput{Username: "user1"}).Return(service.RetrieveUserInfoOutput{}, errors.New("unexpected error"))
			},
			expectedCode: http.StatusInternalServerError,
			expectedBody: `{"errors":"internal error"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := fiber.New()
			r := UserRoutes{log: logger, userService: mockUserService}
			app.Get("/info", func(c *fiber.Ctx) error {
				c.Locals("username", tt.username)
				return r.getInfo(c, ctx)
			})

			tt.mockUserFunc()

			req := httptest.NewRequest(http.MethodGet, "/info", nil)
			resp, _ := app.Test(req)

			assert.Equal(t, tt.expectedCode, resp.StatusCode)

			var bodyBytes bytes.Buffer
			_, _ = bodyBytes.ReadFrom(resp.Body)
			assert.Contains(t, bodyBytes.String(), tt.expectedBody)
		})
	}
}
