package v1

import (
	"context"
	"errors"

	"avito-internship/internal/service"
	"avito-internship/internal/service/servicerrs"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

type UserRoutes struct {
	log         *zap.Logger
	userService service.User
}

func newUserRoutes(ctx context.Context, log *zap.Logger, g *fiber.Router, userService service.User) {
	r := UserRoutes{
		log:         log,
		userService: userService,
	}

	(*g).Get("/info", func(c *fiber.Ctx) error {
		return r.getInfo(c, ctx)
	})
}

type InfoResponse struct {
	Coins       int         `json:"coins"`
	Inventory   []Inventory `json:"inventory"`
	CoinHistory CoinHistory `json:"coinHistory"`
}

type Inventory struct {
	Type     string `json:"type"`
	Quantity int    `json:"quantity"`
}

type CoinHistory struct {
	Received []CoinTransferIn  `json:"received"`
	Sent     []CoinTransferOut `json:"sent"`
}

type CoinTransferIn struct {
	User   string `json:"fromUser,omitempty"`
	Amount int    `json:"amount,omitempty"`
}

type CoinTransferOut struct {
	User   string `json:"toUser,omitempty"`
	Amount int    `json:"amount,omitempty"`
}

func (r *UserRoutes) getInfo(c *fiber.Ctx, ctx context.Context) error {
	const op = "v1.userRoutes.getInfo"

	r.log.Info("attempting to get username from context")
	username, ok := c.Locals("username").(string)
	if !ok {
		r.log.Error("failed to extract username from context",
			zap.String("op", op),
			zap.String("route", "api/info"),
		)

		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"errors": "invalid request body",
		})
	}

	r.log.Info("username successfully extracted")

	info, err := r.userService.RetrieveUserInfo(
		ctx,
		service.RetrieveUserInfoInput{
			Username: username,
		})
	if err != nil {
		if errors.Is(err, servicerrs.ErrUserNotFound) {
			r.log.Warn("failed to get info about user",
				zap.String("op", op),
				zap.String("route", "api/info"),
				zap.String("username", username),
			)

			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"errors": "user not found",
			})
		}

		r.log.Error("failed to get info about user",
			zap.String("op", op),
			zap.String("route", "api/info"),
			zap.String("username", username),
			zap.Error(err),
		)

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"errors": "internal error",
		})
	}

	var inventory []Inventory
	for _, item := range info.Inventory {
		inv := Inventory{
			Type:     item.Product,
			Quantity: item.Quantity,
		}
		inventory = append(inventory, inv)
	}

	var transfersIn []CoinTransferIn
	for _, item := range info.TransferIn {
		transferIn := CoinTransferIn{
			User:   item.Username,
			Amount: item.Amount,
		}
		transfersIn = append(transfersIn, transferIn)
	}

	var transfersOut []CoinTransferOut
	for _, item := range info.TransferOut {
		transferOut := CoinTransferOut{
			User:   item.Username,
			Amount: item.Amount,
		}
		transfersOut = append(transfersOut, transferOut)
	}

	response := InfoResponse{
		Coins:     info.Balance,
		Inventory: inventory,
		CoinHistory: CoinHistory{
			Received: transfersIn,
			Sent:     transfersOut,
		},
	}

	return c.JSON(response)
}
