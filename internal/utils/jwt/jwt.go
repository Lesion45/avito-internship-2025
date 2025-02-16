package jwt

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// NewToken creates new JWT token for given user by his username.
func NewToken(username string, salt string, tokenTTL time.Duration) (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)

	claims := token.Claims.(jwt.MapClaims)
	claims["username"] = username
	claims["exp"] = time.Now().Add(tokenTTL).Unix()

	tokenString, err := token.SignedString([]byte(salt))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// ParseToken extracts username from token.
func ParseToken(tokenString string, secretKey string) (string, error) {
	const op = "ParseToken"

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("%s: unexpected signing method: %v", op, token.Header["alg"])
		}
		return []byte(secretKey), nil
	})

	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}

	if !token.Valid {
		return "", fmt.Errorf("%s: invalid token", op)
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", fmt.Errorf("%s: invalid token claims", op)
	}

	username, ok := claims["username"].(string)
	if !ok {
		return "", fmt.Errorf("%s: email not found in token", op)
	}

	return username, nil
}
