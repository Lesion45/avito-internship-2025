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

type operationRoutes struct {
	log              *zap.Logger
	operationService service.Operation
}

func newOperationRoutes(ctx context.Context, log *zap.Logger, g *fiber.Router, operationService service.Operation) {
	r := operationRoutes{
		log:              log,
		operationService: operationService,
	}

	(*g).Post("/sendCoin", func(c *fiber.Ctx) error {
		return r.sendCoin(c, ctx)
	})

	(*g).Get("/buy/:item", func(c *fiber.Ctx) error {
		return r.buyProduct(c, ctx)
	})
}

type SendCoinRequest struct {
	ToUser string `json:"toUser" validate:"required"`
	Amount int    `json:"amount" validate:"required,gt=0"`
}

func (r operationRoutes) sendCoin(c *fiber.Ctx, ctx context.Context) error {
	const op = "v1.operationService.sendCoin"

	r.log.Info("attempting to get username from context")
	username, ok := c.Locals("username").(string)
	if !ok {
		r.log.Error("failed to extract username from context",
			zap.String("op", op),
			zap.String("route", "api/sendCoin"),
		)

		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"errors": "invalid request body",
		})
	}

	r.log.Info("username successfully extracted")

	r.log.Info("attempting to decode request body")
	var req SendCoinRequest
	if err := c.BodyParser(&req); err != nil {
		r.log.Error("failed to decode request body",
			zap.String("op", op),
			zap.String("route", "api/sendCoin"),
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
			zap.String("route", "api/sendCoin"),
			zap.Error(err),
		)

		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"errors": validation.ValidataionError(validateErr),
		})
	}

	if username == req.ToUser {
		r.log.Warn("Wrong recipient",
			zap.String("op", op),
			zap.String("route", "api/sendCoin"),
			zap.Error(errors.New("you cannot sent coins to yourself")),
		)

		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"errors": "you cannot sent coins to yourself",
		})
	}

	err := r.operationService.TransferFunds(ctx, service.TransferFundsInput{
		Sender:    username,
		Recipient: req.ToUser,
		Amount:    req.Amount,
	})
	if err != nil {
		if errors.Is(err, servicerrs.ErrRecipientNotFound) {
			r.log.Error("recipient not found",
				zap.String("op", op),
				zap.String("route", "api/sendCoin"),
				zap.String("recipient", req.ToUser),
			)

			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"errors": "recipient not found",
			})
		} else if errors.Is(err, servicerrs.ErrInsufficientFunds) {
			r.log.Error("insufficient funds",
				zap.String("op", op),
				zap.String("route", "api/sendCoin"),
				zap.String("sender", username),
			)

			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"errors": "insufficient funds",
			})
		}

		r.log.Error("failed to transfer funds",
			zap.String("op", op),
			zap.String("route", "api/sendCoin"),
			zap.String("sender", username),
			zap.Error(err),
		)

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"errors": "internal error",
		})
	}

	return c.SendStatus(fiber.StatusOK)
}

func (r operationRoutes) buyProduct(c *fiber.Ctx, ctx context.Context) error {
	const op = "v1.operationService.sendCoin"

	r.log.Info("attempting to get username from context")
	username, ok := c.Locals("username").(string)
	if !ok {
		r.log.Error("failed to extract username from context",
			zap.String("op", op),
			zap.String("route", "api/buy"),
		)

		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"errors": "invalid request body",
		})
	}

	r.log.Info("username successfully extracted")

	r.log.Info("attempting to decode request body")
	item := c.Params("item")
	if item == "" {
		r.log.Error("invalid request",
			zap.String("op", op),
			zap.String("route", "api/buy"),
		)

		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"errors": "Item is required",
		})
	}
	r.log.Info("request body decoded")

	err := r.operationService.PurchaseProduct(ctx, service.PurchaseProductInput{
		Username: username,
		Product:  item,
	})
	if err != nil {
		if errors.Is(err, servicerrs.ErrInsufficientFunds) {
			r.log.Error("insufficient funds",
				zap.String("op", op),
				zap.String("route", "api/buy"),
				zap.String("customer", username),
			)

			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"errors": "insufficient funds",
			})
		} else if errors.Is(err, servicerrs.ErrProductNotFound) {
			r.log.Error("product not found",
				zap.String("op", op),
				zap.String("route", "api/buy"),
				zap.String("customer", username),
			)

			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"errors": "product not found",
			})
		}

		r.log.Error("failed to buy product",
			zap.String("op", op),
			zap.String("route", "api/auth"),
			zap.String("customer", username),
			zap.String("item", item),
			zap.Error(err),
		)

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"errors": "internal error",
		})
	}

	return c.SendStatus(fiber.StatusOK)
}
