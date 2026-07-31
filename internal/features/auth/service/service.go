package auth_service

import (
	"context"
	"discord/internal/core/domain"
	"time"
)

type AuthService struct {
	authRepo     AuthRepository
	usersService UserService
}

type AuthRepository interface {
	Transaction(ctx context.Context, fn func(ctx context.Context) error) error
	CreateSession(ctx context.Context, userID int64, token string, expiresAt time.Time) (domain.Session, error)
	GetSessionByToken(ctx context.Context, token string) (domain.Session, error)
	DeleteSession(ctx context.Context, token string) error
}

// UserService — интерфейс для user/service, чтобы AuthService
// вызывал сервисный слой, а не репозиторий напрямую.
type UserService interface {
	SaveUser(ctx context.Context, user domain.User) (domain.User, error)
	GetUser(ctx context.Context, userID *int) (domain.User, error)
	GetUserID(ctx context.Context, username string) (int64, error)
}

func NewAuthService(authRepo AuthRepository, usersService UserService) *AuthService {
	return &AuthService{
		authRepo:     authRepo,
		usersService: usersService,
	}
}
