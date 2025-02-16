package jwt

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
)

func Test_ParseToken(t *testing.T) {
	secretKey := "supersecretkey"

	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)
	claims["username"] = "test_username"
	claims["exp"] = time.Now().Add(24 * time.Hour).Unix()
	tokenString, err := token.SignedString([]byte(secretKey))
	assert.NoError(t, err)

	tests := []struct {
		name         string
		tokenString  string
		secretKey    string
		wantUsername string
		wantErr      bool
	}{
		{
			name:         "Valid token",
			tokenString:  tokenString,
			secretKey:    secretKey,
			wantUsername: "test_username",
			wantErr:      false,
		},
		{
			name:         "Invalid signature",
			tokenString:  tokenString,
			secretKey:    "wrongkey",
			wantUsername: "",
			wantErr:      true,
		},
		{
			name:         "Invalid token format",
			tokenString:  "invalid.token.format",
			secretKey:    secretKey,
			wantUsername: "",
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			email, err := ParseToken(tt.tokenString, tt.secretKey)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.wantUsername, email)
		})
	}
}
