package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"avito-internship/internal/repository"
	"avito-internship/internal/repository/repoerrs"
	"avito-internship/internal/service/servicerrs"
	"avito-internship/internal/utils/jwt"

	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	log            *zap.Logger
	userRepository repository.User
	tokenTTL       time.Duration
	salt           string
}

func NewAuthService(log *zap.Logger, repo repository.User, tokenTTL time.Duration, salt string) *AuthService {
	return &AuthService{
		log:            log,
		userRepository: repo,
		tokenTTL:       tokenTTL,
		salt:           salt,
	}
}

func (s *AuthService) Authorization(ctx context.Context, username, password string) (string, error) {
	const op = "service.Auth.Authorization"
	s.log.Info("Attempting to authorize user", zap.String("username", username))

	user, err := s.userRepository.GetUserCredentials(ctx, username)
	if err != nil {
		if errors.Is(err, repoerrs.ErrUserNotFound) {
			return s.handleUserNotFound(ctx, username, password, op)
		}
		s.log.Error("Failed to retrieve user credentials",
			zap.String("op", op),
			zap.Error(err),
		)
		return "", fmt.Errorf("%s: %w", op, err)
	}

	if err := bcrypt.CompareHashAndPassword(user.Password, []byte(password)); err != nil {
		s.log.Warn("Invalid credentials",
			zap.String("op", op),
			zap.String("username", username),
			zap.Error(err),
		)
		return "", fmt.Errorf("%s: %w", op, servicerrs.ErrInvalidCredentials)
	}

	return s.generateToken(username, op)
}

func (s *AuthService) handleUserNotFound(ctx context.Context, username, password, op string) (string, error) {
	s.log.Info("User not found, creating a new one", zap.String("username", username))

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		s.log.Error("Failed to generate password hash",
			zap.String("op", op),
			zap.Error(err),
		)
		return "", fmt.Errorf("%s: %w", op, err)
	}

	if err := s.userRepository.AddUser(ctx, username, passwordHash); err != nil {
		s.log.Error("Failed to create user",
			zap.String("op", op),
			zap.Error(err),
		)
		return "", fmt.Errorf("%s: %w", op, err)
	}

	return s.generateToken(username, op)
}

func (s *AuthService) generateToken(username, op string) (string, error) {
	token, err := jwt.NewToken(username, s.salt, s.tokenTTL)
	if err != nil {
		s.log.Error("Failed to generate token",
			zap.String("op", op),
			zap.String("username", username),
			zap.Error(err),
		)
		return "", fmt.Errorf("%s: %w", op, err)
	}

	s.log.Info("User successfully authorized",
		zap.String("username", username),
	)

	return token, nil
}

func (s *AuthService) ValidateToken(token string) (string, error) {
	const op = "service.Auth.ValidateToken"

	if len(token) > 7 && strings.HasPrefix(token, "Bearer ") {
		token = token[7:]
	}

	username, err := jwt.ParseToken(token, s.salt)
	if err != nil {
		s.log.Warn("Invalid token",
			zap.String("op", op),
			zap.Error(err),
		)
		return "", fmt.Errorf("%s: %w", op, servicerrs.ErrInvalidToken)
	}

	s.log.Info("Token validated successfully",
		zap.String("username", username),
	)

	return username, nil
}
