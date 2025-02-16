package v1

import (
	"context"
	"errors"

	"avito-internship/internal/service"
	"avito-internship/internal/service/servicerrs"
	"avito-internship/pkg/validation"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

type AuthRoutes struct {
	log         *zap.Logger
	authService service.Auth
}

func newAuthRoutes(ctx context.Context, log *zap.Logger, g *fiber.Router, authService service.Auth) {
	r := AuthRoutes{
		log:         log,
		authService: authService,
	}

	(*g).Post("/auth", func(c *fiber.Ctx) error {
		return r.authorize(c, ctx)
	})
}

type AuthRequest struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}

type AuthResponse struct {
	Token string `json:"token"`
}

func (r *AuthRoutes) authorize(c *fiber.Ctx, ctx context.Context) error {
	const op = "v1.authRoutes.authorize"

	var req AuthRequest

	r.log.Info("attempting to decode request body")
	if err := c.BodyParser(&req); err != nil {
		r.log.Error("failed to decode request body",
			zap.String("op", op),
			zap.String("route", "api/auth"),
			zap.Error(err),
		)

		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"errors": "invalid request body",
		})
	}

	r.log.Info("request body decoded")
	if err := validator.New().Struct(req); err != nil {
		var validateErr validator.ValidationErrors
		errors.As(err, &validateErr)
		r.log.Error("invalid request",
			zap.String("op", op),
			zap.Error(err),
		)

		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"errors": validation.ValidataionError(validateErr),
		})
	}

	token, err := r.authService.Authorization(ctx, req.Username, req.Password)
	if err != nil {
		if errors.Is(err, servicerrs.ErrInvalidCredentials) {
			r.log.Warn("invalid credentials",
				zap.String("op", op),
				zap.String("route", "api/auth"),
				zap.String("username", req.Username),
				zap.Error(err),
			)

			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"errors": "invalid credentials",
			})
		}

		r.log.Error("failed to authorize user",
			zap.String("op", op),
			zap.String("route", "api/auth"),
			zap.String("username", req.Username),
			zap.Error(err),
		)

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"errors": "internal error",
		})
	}

	response := AuthResponse{
		Token: token,
	}

	return c.JSON(response)
}
