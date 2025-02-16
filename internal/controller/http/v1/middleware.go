package v1

import (
	"avito-internship/internal/service"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

type AuthMiddleware struct {
	authService service.Auth
	log         *zap.Logger
}

func NewAuthMiddleware(log *zap.Logger, authService service.Auth) *AuthMiddleware {
	return &AuthMiddleware{
		authService: authService,
		log:         log,
	}
}

func (m *AuthMiddleware) Auth() fiber.Handler {
	return func(c *fiber.Ctx) error {
		const op = "middleware.AuthMiddleware"

		token := c.Get("Authorization")
		if token == "" {
			m.log.Warn("Missing authorization token", zap.String("op", op))
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized"})
		}

		username, err := m.authService.ValidateToken(token)
		if err != nil {
			m.log.Warn("Invalid token", zap.String("op", op), zap.Error(err))
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid token"})
		}

		c.Locals("username", username)

		return c.Next()
	}
}
