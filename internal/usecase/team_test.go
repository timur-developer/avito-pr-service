package usecase

import (
	"avito-pr-service/internal/models"
	"avito-pr-service/internal/repository"
	"context"
	"errors"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"io"
	"log/slog"
	"testing"
)

type mockTeamRepository struct{ mock.Mock }

func (m *mockTeamRepository) CreateTeam(ctx context.Context, team models.Team) error {
	return m.Called(ctx, team).Error(0)
}

func (m *mockTeamRepository) GetTeam(ctx context.Context, name string) (models.Team, error) {
	args := m.Called(ctx, name)
	return args.Get(0).(models.Team), args.Error(1)
}

func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func TestTeamUsecase_AddTeam_AlreadyExists(t *testing.T) {
	repo := new(mockTeamRepository)
	uc := NewTeamUsecase(repo, testLogger())

	repo.On("GetTeam", mock.Anything, "avito").Return(models.Team{}, nil)

	err := uc.AddTeam(context.Background(), models.Team{Name: "avito"})
	var appErr models.AppError
	require.True(t, errors.As(err, &appErr))
	require.Equal(t, models.ErrorTeamExists, appErr.Code)
	require.Equal(t, "team already exists", appErr.Message)

	repo.AssertExpectations(t)
}

func TestTeamUsecase_AddTeam_Success(t *testing.T) {
	repo := new(mockTeamRepository)
	uc := NewTeamUsecase(repo, testLogger())

	repo.On("GetTeam", mock.Anything, "new-team").Return(models.Team{}, repository.ErrNotFound)
	repo.On("CreateTeam", mock.Anything, mock.Anything).Return(nil)

	err := uc.AddTeam(context.Background(), models.Team{Name: "new-team"})
	require.NoError(t, err)

	repo.AssertExpectations(t)
}

func TestTeamUsecase_GetTeam_NotFound(t *testing.T) {
	repo := new(mockTeamRepository)
	uc := NewTeamUsecase(repo, testLogger())

	repo.On("GetTeam", mock.Anything, "unknown").Return(models.Team{}, repository.ErrNotFound)

	_, err := uc.GetTeam(context.Background(), "unknown")
	var appErr models.AppError
	require.True(t, errors.As(err, &appErr))
	require.Equal(t, models.ErrorNotFound, appErr.Code)

	repo.AssertExpectations(t)
}

func TestTeamUsecase_GetTeam_Success(t *testing.T) {
	repo := new(mockTeamRepository)
	uc := NewTeamUsecase(repo, testLogger())

	expected := models.Team{Name: "avito", Members: []models.TeamMember{{UserID: "u1"}}}
	repo.On("GetTeam", mock.Anything, "avito").Return(expected, nil)

	got, err := uc.GetTeam(context.Background(), "avito")
	require.NoError(t, err)
	require.Equal(t, expected.Name, got.Name)
	require.Len(t, got.Members, 1)

	repo.AssertExpectations(t)
}
