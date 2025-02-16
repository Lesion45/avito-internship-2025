package servicerrs

import "errors"

var (
	ErrRecipientNotFound  = errors.New("recipient not found")
	ErrInsufficientFunds  = errors.New("insufficient funds")
	ErrCustomerNotFound   = errors.New("customer not found")
	ErrProductNotFound    = errors.New("product not found")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrInvalidToken       = errors.New("invalid token")
	ErrUserNotFound       = errors.New("user not found")
	ErrUserAlreadyExists  = errors.New("user already exists")
)
