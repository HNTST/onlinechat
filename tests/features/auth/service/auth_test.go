package auth_service_test

import (
	"context"
	"discord/internal/core/domain"
	core_errors "discord/internal/core/errors"
	"errors"
	"fmt"
	"testing"
	"time"

	auth_service "discord/internal/features/auth/service"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

type mockAuthRepo struct {
	mock.Mock
}

func (m *mockAuthRepo) Transaction(ctx context.Context, fn func(ctx context.Context) error) error {
	args := m.Called(ctx, fn)
	return args.Error(0)
}

func (m *mockAuthRepo) CreateSession(ctx context.Context, userID int64, token string, expiresAt time.Time) (domain.Session, error) {
	args := m.Called(ctx, userID, token, expiresAt)
	return args.Get(0).(domain.Session), args.Error(1)
}

func (m *mockAuthRepo) GetSessionByToken(ctx context.Context, token string) (domain.Session, error) {
	args := m.Called(ctx, token)
	return args.Get(0).(domain.Session), args.Error(1)
}

func (m *mockAuthRepo) DeleteSession(ctx context.Context, token string) error {
	args := m.Called(ctx, token)
	return args.Error(0)
}

type mockUsersRepo struct {
	mock.Mock
}

func (m *mockUsersRepo) GetUser(ctx context.Context, userID int64) (domain.User, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(domain.User), args.Error(1)
}

func (m *mockUsersRepo) GetUserID(ctx context.Context, username string) (int64, error) {
	args := m.Called(ctx, username)
	return args.Get(0).(int64), args.Error(1)
}

func hashPassword(t *testing.T, password string) string {
	t.Helper()
	h, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	require.NoError(t, err)
	return string(h)
}

func TestLogin_Success(t *testing.T) {
	mockAuth := new(mockAuthRepo)
	mockUsers := new(mockUsersRepo)
	svc := auth_service.NewAuthService(mockAuth, mockUsers)
	ctx := context.Background()

	password := "secret123"
	hashed := hashPassword(t, password)
	user := domain.NewUserWithAvatar(1, "alice", hashed, time.Now(), nil)

	mockUsers.On("GetUserID", ctx, "alice").Return(int64(1), nil)
	mockUsers.On("GetUser", ctx, int64(1)).Return(user, nil)
	mockAuth.On("CreateSession", ctx, int64(1), mock.AnythingOfType("string"), mock.AnythingOfType("time.Time")).
		Return(domain.Session{ID: 1, UserID: 1, Token: "test-token", ExpiresAt: time.Now().Add(time.Hour)}, nil)

	token, u, err := svc.Login(ctx, "alice", password)

	require.NoError(t, err)
	require.Equal(t, "test-token", token)
	require.Equal(t, "alice", u.Name)
	mockAuth.AssertExpectations(t)
	mockUsers.AssertExpectations(t)
}

func TestLogin_WrongPassword(t *testing.T) {
	mockAuth := new(mockAuthRepo)
	mockUsers := new(mockUsersRepo)
	svc := auth_service.NewAuthService(mockAuth, mockUsers)
	ctx := context.Background()

	hashed := hashPassword(t, "correctpassword")
	user := domain.NewUserWithAvatar(1, "alice", hashed, time.Now(), nil)

	mockUsers.On("GetUserID", ctx, "alice").Return(int64(1), nil)
	mockUsers.On("GetUser", ctx, int64(1)).Return(user, nil)

	_, _, err := svc.Login(ctx, "alice", "wrongpassword")

	require.Error(t, err)
	require.True(t, errors.Is(err, core_errors.ErrUnauthorized))
	mockUsers.AssertExpectations(t)
	mockAuth.AssertNotCalled(t, "CreateSession")
}

func TestLogin_UserNotFound(t *testing.T) {
	mockAuth := new(mockAuthRepo)
	mockUsers := new(mockUsersRepo)
	svc := auth_service.NewAuthService(mockAuth, mockUsers)
	ctx := context.Background()

	mockUsers.On("GetUserID", ctx, "nonexistent").Return(int64(0), fmt.Errorf("user not found: %w", core_errors.ErrUserNotFound))

	_, _, err := svc.Login(ctx, "nonexistent", "password")

	require.Error(t, err)
	require.True(t, errors.Is(err, core_errors.ErrUnauthorized))
	mockUsers.AssertExpectations(t)
	mockAuth.AssertNotCalled(t, "CreateSession")
}

func TestLogout_Success(t *testing.T) {
	mockAuth := new(mockAuthRepo)
	mockUsers := new(mockUsersRepo)
	svc := auth_service.NewAuthService(mockAuth, mockUsers)
	ctx := context.Background()

	mockAuth.On("DeleteSession", ctx, "test-token").Return(nil)

	err := svc.Logout(ctx, "test-token")

	require.NoError(t, err)
	mockAuth.AssertExpectations(t)
}

func TestValidateSession_Success(t *testing.T) {
	mockAuth := new(mockAuthRepo)
	mockUsers := new(mockUsersRepo)
	svc := auth_service.NewAuthService(mockAuth, mockUsers)
	ctx := context.Background()

	session := domain.Session{
		ID:        1,
		UserID:    42,
		Token:     "valid-token",
		ExpiresAt: time.Now().UTC().Add(time.Hour),
	}

	mockAuth.On("GetSessionByToken", ctx, "valid-token").Return(session, nil)

	userID, err := svc.ValidateSession(ctx, "valid-token")

	require.NoError(t, err)
	require.Equal(t, int64(42), userID)
	mockAuth.AssertExpectations(t)
}

func TestValidateSession_Expired(t *testing.T) {
	mockAuth := new(mockAuthRepo)
	mockUsers := new(mockUsersRepo)
	svc := auth_service.NewAuthService(mockAuth, mockUsers)
	ctx := context.Background()

	session := domain.Session{
		ID:        1,
		UserID:    42,
		Token:     "expired-token",
		ExpiresAt: time.Now().UTC().Add(-time.Hour),
	}

	mockAuth.On("GetSessionByToken", ctx, "expired-token").Return(session, nil)
	mockAuth.On("DeleteSession", ctx, "expired-token").Return(nil)

	_, err := svc.ValidateSession(ctx, "expired-token")

	require.Error(t, err)
	require.True(t, errors.Is(err, core_errors.ErrUnauthorized))
	mockAuth.AssertExpectations(t)
}

func TestValidateSession_TokenNotFound(t *testing.T) {
	mockAuth := new(mockAuthRepo)
	mockUsers := new(mockUsersRepo)
	svc := auth_service.NewAuthService(mockAuth, mockUsers)
	ctx := context.Background()

	mockAuth.On("GetSessionByToken", ctx, "invalid-token").Return(domain.Session{}, fmt.Errorf("session not found: %w", core_errors.ErrUnauthorized))

	_, err := svc.ValidateSession(ctx, "invalid-token")

	require.Error(t, err)
	require.True(t, errors.Is(err, core_errors.ErrUnauthorized))
	mockAuth.AssertExpectations(t)
}
