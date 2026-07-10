package auth_service

import (
	"context"
	"discord/internal/core/domain"
	"time"
)

type AuthService struct {
	authRepo    AuthRepository
	usersRepo   UsersRepository
}

type AuthRepository interface {
	Transaction(ctx context.Context, fn func(ctx context.Context) error) error
	CreateSession(ctx context.Context, userID int64, token string, expiresAt time.Time) (domain.Session, error)
	GetSessionByToken(ctx context.Context, token string) (domain.Session, error)
	DeleteSession(ctx context.Context, token string) error
}

type UsersRepository interface {
	SaveUser(ctx context.Context, user domain.User) (domain.User, error)
	GetUser(ctx context.Context, userID int64) (domain.User, error)
	GetUserID(ctx context.Context, username string) (int64, error)
}

func NewAuthService(authRepo AuthRepository, usersRepo UsersRepository) *AuthService {
	return &AuthService{
		authRepo:  authRepo,
		usersRepo: usersRepo,
	}
}
