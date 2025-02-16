package v1

import (
	"context"

	"avito-internship/internal/service"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

func InitRouter(ctx context.Context, log *zap.Logger, app *fiber.App, services *service.Services, middleware fiber.Handler) {
	v1 := app.Group("api")

	// Public routes
	newAuthRoutes(ctx, log, &v1, services.Auth)

	// Protected with auth middleware
	protected := v1.Group("")
	protected.Use(middleware)

	newUserRoutes(ctx, log, &protected, services.User)
	newOperationRoutes(ctx, log, &protected, services.Operation)
}
