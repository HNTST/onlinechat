package users_service_test

import (
	"context"
	"discord/internal/core/domain"
	"time"

	"github.com/stretchr/testify/mock"
)

type mockUserRepo struct {
	mock.Mock
}

func (m *mockUserRepo) SaveUser(ctx context.Context, user domain.User) (domain.User, error) {
	args := m.Called(ctx, user)
	return args.Get(0).(domain.User), args.Error(1)
}

func (m *mockUserRepo) GetUser(ctx context.Context, userID int64) (domain.User, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(domain.User), args.Error(1)
}

func (m *mockUserRepo) GetUserID(ctx context.Context, username string) (int64, error) {
	args := m.Called(ctx, username)
	return args.Get(0).(int64), args.Error(1)
}

func (m *mockUserRepo) SaveLastSeen(ctx context.Context, userID int64, lastSeen time.Time) error {
	args := m.Called(ctx, userID, lastSeen)
	return args.Error(0)
}

func (m *mockUserRepo) GetLastSeen(ctx context.Context, userID int64) (*time.Time, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*time.Time), args.Error(1)
}

func (m *mockUserRepo) UpdateAvatar(ctx context.Context, userID, fileID int64) error {
	args := m.Called(ctx, userID, fileID)
	return args.Error(0)
}
