package usecase

import (
	"avito-pr-service/internal/models"
	"context"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"testing"
)

type mockUserRepository struct{ mock.Mock }

func (m *mockUserRepository) SetActive(ctx context.Context, userID string, isActive bool) error {
	return m.Called(ctx, userID, isActive).Error(0)
}

func (m *mockUserRepository) GetUser(ctx context.Context, userID string) (models.User, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(models.User), args.Error(1)
}
func TestUserUsecase_SetActive_Success(t *testing.T) {
	repo := new(mockUserRepository)
	uc := NewUserUsecase(repo, testLogger())

	repo.On("SetActive", mock.Anything, "u1", false).Return(nil)

	err := uc.SetActive(context.Background(), models.SetUserActiveRequest{
		UserID: "u1", IsActive: false,
	})

	require.NoError(t, err)
	repo.AssertExpectations(t)
}
