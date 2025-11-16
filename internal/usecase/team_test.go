package usecase

import (
	"avito-pr-service/internal/models"
	"context"
	"errors"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
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

func TestTeamUsecase_AddTeam_AlreadyExists(t *testing.T) {
	repo := new(mockTeamRepository)
	userRepo := new(mockUserRepository)
	prRepo := new(mockPRRepository)
	prUC := NewPRUsecase(prRepo, userRepo, repo, testLogger())

	uc := NewTeamUsecase(repo, userRepo, prRepo, prUC, testLogger())

	team := models.Team{
		Name: "avito",
		Members: []models.TeamMember{
			{UserID: "u1", Username: "Alice", IsActive: true},
		},
	}

	userRepo.On("GetUser", mock.Anything, "u1").Return(models.User{}, models.ErrUserNotFound)

	repo.On("GetTeam", mock.Anything, "avito").Return(models.Team{}, nil)

	err := uc.AddTeam(context.Background(), team)
	var appErr models.AppError
	require.True(t, errors.As(err, &appErr))
	require.Equal(t, models.ErrorTeamExists, appErr.Code)
	require.Equal(t, "team already exists", appErr.Message)

	repo.AssertExpectations(t)
	userRepo.AssertExpectations(t)
}

func TestTeamUsecase_AddTeam_Success(t *testing.T) {
	repo := new(mockTeamRepository)
	userRepo := new(mockUserRepository)
	prRepo := new(mockPRRepository)
	prUC := NewPRUsecase(prRepo, userRepo, repo, testLogger())

	uc := NewTeamUsecase(repo, userRepo, prRepo, prUC, testLogger())

	team := models.Team{
		Name: "new-team",
		Members: []models.TeamMember{
			{UserID: "u1", Username: "Alice", IsActive: true},
			{UserID: "u2", Username: "Bob", IsActive: true},
		},
	}

	userRepo.On("GetUser", mock.Anything, "u1").Return(models.User{}, models.ErrUserNotFound)
	userRepo.On("GetUser", mock.Anything, "u2").Return(models.User{}, models.ErrUserNotFound)

	repo.On("GetTeam", mock.Anything, "new-team").Return(models.Team{}, models.ErrTeamNotFound)
	repo.On("CreateTeam", mock.Anything, mock.MatchedBy(func(t models.Team) bool {
		return t.Name == "new-team" && len(t.Members) == 2
	})).Return(nil)

	err := uc.AddTeam(context.Background(), team)
	require.NoError(t, err)

	repo.AssertExpectations(t)
	userRepo.AssertExpectations(t)
}

func TestTeamUsecase_AddTeam_EmptyMembers(t *testing.T) {
	repo := new(mockTeamRepository)
	userRepo := new(mockUserRepository)
	prRepo := new(mockPRRepository)
	prUC := NewPRUsecase(prRepo, userRepo, repo, testLogger())

	uc := NewTeamUsecase(repo, userRepo, prRepo, prUC, testLogger())

	team := models.Team{
		Name:    "empty-team",
		Members: []models.TeamMember{},
	}

	err := uc.AddTeam(context.Background(), team)
	var appErr models.AppError
	require.True(t, errors.As(err, &appErr))
	require.Equal(t, models.ErrorEmptyTeam, appErr.Code)
}

func TestTeamUsecase_AddTeam_DuplicateUserID(t *testing.T) {
	repo := new(mockTeamRepository)
	userRepo := new(mockUserRepository)
	prRepo := new(mockPRRepository)
	prUC := NewPRUsecase(prRepo, userRepo, repo, testLogger())

	uc := NewTeamUsecase(repo, userRepo, prRepo, prUC, testLogger())

	team := models.Team{
		Name: "duplicate-team",
		Members: []models.TeamMember{
			{UserID: "u1", Username: "Alice", IsActive: true},
			{UserID: "u1", Username: "Alice2", IsActive: true},
		},
	}

	err := uc.AddTeam(context.Background(), team)
	var appErr models.AppError
	require.True(t, errors.As(err, &appErr))
	require.Equal(t, models.ErrorDuplicateUserID, appErr.Code)
}

func TestTeamUsecase_AddTeam_UserInAnotherTeam(t *testing.T) {
	repo := new(mockTeamRepository)
	userRepo := new(mockUserRepository)
	prRepo := new(mockPRRepository)
	prUC := NewPRUsecase(prRepo, userRepo, repo, testLogger())

	uc := NewTeamUsecase(repo, userRepo, prRepo, prUC, testLogger())

	team := models.Team{
		Name: "new-team",
		Members: []models.TeamMember{
			{UserID: "u1", Username: "Alice", IsActive: true},
		},
	}

	existingUser := models.User{
		UserID:   "u1",
		Username: "Alice",
		TeamName: "existing-team",
		IsActive: true,
	}
	userRepo.On("GetUser", mock.Anything, "u1").Return(existingUser, nil)

	err := uc.AddTeam(context.Background(), team)
	var appErr models.AppError
	require.True(t, errors.As(err, &appErr))
	require.Equal(t, models.ErrorUserInAnotherTeam, appErr.Code)

	userRepo.AssertExpectations(t)
}

func TestTeamUsecase_GetTeam_NotFound(t *testing.T) {
	repo := new(mockTeamRepository)
	userRepo := new(mockUserRepository)
	prRepo := new(mockPRRepository)
	prUC := NewPRUsecase(prRepo, userRepo, repo, testLogger())

	uc := NewTeamUsecase(repo, userRepo, prRepo, prUC, testLogger())

	repo.On("GetTeam", mock.Anything, "unknown").Return(models.Team{}, models.ErrTeamNotFound)

	_, err := uc.GetTeam(context.Background(), "unknown")
	var appErr models.AppError
	require.True(t, errors.As(err, &appErr))
	require.Equal(t, models.ErrorNotFound, appErr.Code)

	repo.AssertExpectations(t)
}

func TestTeamUsecase_GetTeam_Success(t *testing.T) {
	repo := new(mockTeamRepository)
	userRepo := new(mockUserRepository)
	prRepo := new(mockPRRepository)
	prUC := NewPRUsecase(prRepo, userRepo, repo, testLogger())

	uc := NewTeamUsecase(repo, userRepo, prRepo, prUC, testLogger())

	expected := models.Team{
		Name: "avito",
		Members: []models.TeamMember{
			{UserID: "u1", Username: "Alice", IsActive: true},
		},
	}
	repo.On("GetTeam", mock.Anything, "avito").Return(expected, nil)

	got, err := uc.GetTeam(context.Background(), "avito")
	require.NoError(t, err)
	require.Equal(t, expected.Name, got.Name)
	require.Len(t, got.Members, 1)

	repo.AssertExpectations(t)
}
