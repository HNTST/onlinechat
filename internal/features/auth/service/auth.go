package auth_service

import (
	"context"
	"crypto/rand"
	"discord/internal/core/domain"
	core_errors "discord/internal/core/errors"
	"encoding/hex"
	"fmt"
	"time"

	"golang.org/x/crypto/bcrypt"
)

const sessionTokenBytes = 32
const sessionTTL = 24 * time.Hour

func generateSessionToken() (string, error) {
	b := make([]byte, sessionTokenBytes)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generate token: %w", err)
	}
	return hex.EncodeToString(b), nil
}

func (s *AuthService) Login(ctx context.Context, username, password string) (string, domain.User, error) {
	userID, err := s.usersRepo.GetUserID(ctx, username)
	if err != nil {
		return "", domain.User{}, fmt.Errorf("login: %w", core_errors.ErrUnauthorized)
	}

	user, err := s.usersRepo.GetUser(ctx, userID)
	if err != nil {
		return "", domain.User{}, fmt.Errorf("login: %w", core_errors.ErrUnauthorized)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return "", domain.User{}, fmt.Errorf("login: %w", core_errors.ErrUnauthorized)
	}

	token, err := generateSessionToken()
	if err != nil {
		return "", domain.User{}, fmt.Errorf("login: %w", err)
	}

	expiresAt := time.Now().UTC().Add(sessionTTL)

	session, err := s.authRepo.CreateSession(ctx, user.Id, token, expiresAt)
	if err != nil {
		return "", domain.User{}, fmt.Errorf("login: %w", err)
	}

	return session.Token, user, nil
}

func (s *AuthService) Register(ctx context.Context, username, password string) (string, domain.User, error) {
	user := domain.NewUserUninitialized(username, password)
	if err := user.Validate(); err != nil {
		return "", domain.User{}, fmt.Errorf("register: %w", core_errors.ErrInvalidArgument)
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", domain.User{}, fmt.Errorf("register: %w", err)
	}
	user.Password = string(hashedPassword)

	savedUser, err := s.usersRepo.SaveUser(ctx, user)
	if err != nil {
		return "", domain.User{}, fmt.Errorf("register: %w", err)
	}

	token, err := generateSessionToken()
	if err != nil {
		return "", domain.User{}, fmt.Errorf("register: %w", err)
	}

	expiresAt := time.Now().UTC().Add(sessionTTL)

	session, err := s.authRepo.CreateSession(ctx, savedUser.Id, token, expiresAt)
	if err != nil {
		return "", domain.User{}, fmt.Errorf("register: %w", err)
	}

	return session.Token, savedUser, nil
}

func (s *AuthService) Logout(ctx context.Context, token string) error {
	return s.authRepo.DeleteSession(ctx, token)
}

func (s *AuthService) ValidateSession(ctx context.Context, token string) (int64, error) {
	session, err := s.authRepo.GetSessionByToken(ctx, token)
	if err != nil {
		return 0, fmt.Errorf("validate session: %w", err)
	}
	if time.Now().UTC().After(session.ExpiresAt) {
		_ = s.authRepo.DeleteSession(ctx, token)
		return 0, fmt.Errorf("validate session: %w", core_errors.ErrUnauthorized)
	}
	return session.UserID, nil
}
