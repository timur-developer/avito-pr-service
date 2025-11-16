package usecase

import (
	"avito-pr-service/internal/models"
	"avito-pr-service/internal/utils"
	"context"
	"github.com/stretchr/testify/require"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
)

type mockPRRepository struct {
	mock.Mock
}

func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func (m *mockPRRepository) CreatePR(ctx context.Context, pr models.PullRequest) error {
	return m.Called(ctx, pr).Error(0)
}

func (m *mockPRRepository) GetPR(ctx context.Context, prID string) (models.PullRequest, error) {
	args := m.Called(ctx, prID)
	return args.Get(0).(models.PullRequest), args.Error(1)
}

func (m *mockPRRepository) MergePR(ctx context.Context, prID string) (*time.Time, error) {
	args := m.Called(ctx, prID)
	return args.Get(0).(*time.Time), args.Error(1)
}

func (m *mockPRRepository) ReassignReviewer(ctx context.Context, prID, oldUID, newUID string) error {
	return m.Called(ctx, prID, oldUID, newUID).Error(0)
}

func (m *mockPRRepository) GetPRsByReviewer(ctx context.Context, userID string) ([]models.PullRequest, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]models.PullRequest), args.Error(1)
}

func (m *mockPRRepository) GetOpenPRsWithTeamReviewers(ctx context.Context, teamName string) ([]models.PullRequest, error) {
	args := m.Called(ctx, teamName)
	return args.Get(0).([]models.PullRequest), args.Error(1)
}

func (m *mockPRRepository) GetUserStats(ctx context.Context) ([]models.UserStats, error) {
	args := m.Called(ctx)
	return args.Get(0).([]models.UserStats), args.Error(1)
}

func TestPRUsecase_CreatePR_Success(t *testing.T) {
	prRepo := new(mockPRRepository)
	userRepo := new(mockUserRepository)
	teamRepo := new(mockTeamRepository)

	author := models.User{UserID: "u1", Username: "alice", TeamName: "backend", IsActive: true}
	team := models.Team{
		Name: "backend",
		Members: []models.TeamMember{
			{UserID: "u1", Username: "alice", IsActive: true},
			{UserID: "u2", Username: "bob", IsActive: true},
			{UserID: "u3", Username: "charlie", IsActive: true},
		},
	}

	userRepo.On("GetUser", mock.Anything, "u1").Return(author, nil)
	teamRepo.On("GetTeam", mock.Anything, "backend").Return(team, nil)
	prRepo.On("CreatePR", mock.Anything, mock.MatchedBy(func(pr models.PullRequest) bool {
		return pr.ID == "pr-1001" && pr.Name == "Add search" && len(pr.AssignedReviewers) == 2
	})).Return(nil)

	uc := NewPRUsecase(prRepo, userRepo, teamRepo, testLogger())

	req := models.CreatePRRequest{ID: "pr-1001", Name: "Add search", AuthorID: "u1"}
	pr, err := uc.CreatePR(context.Background(), req)

	require.NoError(t, err)
	require.Equal(t, "pr-1001", pr.ID)
	require.Len(t, pr.AssignedReviewers, 2)

	prRepo.AssertExpectations(t)
	userRepo.AssertExpectations(t)
	teamRepo.AssertExpectations(t)
}

func TestPRUsecase_MergePR_Success(t *testing.T) {
	prRepo := new(mockPRRepository)
	userRepo := new(mockUserRepository)
	teamRepo := new(mockTeamRepository)

	mergedAt := time.Now().UTC().Truncate(time.Millisecond)

	pr := models.PullRequest{
		ID:                "pr-1001",
		Name:              "Fix",
		AuthorID:          "u1",
		Status:            models.StatusOpen,
		AssignedReviewers: []string{"u2"},
		CreatedAt:         utils.Ptr(time.Now()),
	}

	prRepo.On("GetPR", mock.Anything, "pr-1001").Return(pr, nil)
	prRepo.On("MergePR", mock.Anything, "pr-1001").Return(&mergedAt, nil)

	uc := NewPRUsecase(prRepo, userRepo, teamRepo, testLogger())

	result, err := uc.MergePR(context.Background(), "pr-1001")

	require.NoError(t, err)
	require.Equal(t, models.StatusMerged, result.Status)
	require.NotNil(t, result.MergedAt)
	require.Equal(t, &mergedAt, result.MergedAt)

	prRepo.AssertExpectations(t)
	userRepo.AssertExpectations(t)
	teamRepo.AssertExpectations(t)
}

func TestPRUsecase_MergePR_AlreadyMerged(t *testing.T) {
	prRepo := new(mockPRRepository)
	userRepo := new(mockUserRepository)
	teamRepo := new(mockTeamRepository)

	mergedAt := time.Now().UTC().Truncate(time.Millisecond)

	pr := models.PullRequest{
		ID:       "pr-1001",
		Status:   models.StatusMerged,
		MergedAt: &mergedAt,
	}

	prRepo.On("GetPR", mock.Anything, "pr-1001").Return(pr, nil)

	uc := NewPRUsecase(prRepo, userRepo, teamRepo, testLogger())

	result, err := uc.MergePR(context.Background(), "pr-1001")

	require.NoError(t, err)
	require.Equal(t, models.StatusMerged, result.Status)
	require.Equal(t, &mergedAt, result.MergedAt)

	prRepo.AssertExpectations(t)
	userRepo.AssertExpectations(t)
	teamRepo.AssertExpectations(t)
}

func TestPRUsecase_ReassignReviewer_Success(t *testing.T) {
	prRepo := new(mockPRRepository)
	userRepo := new(mockUserRepository)
	teamRepo := new(mockTeamRepository)

	pr := models.PullRequest{
		ID:                "pr-1001",
		Status:            models.StatusOpen,
		AssignedReviewers: []string{"u2"},
	}

	updatedPR := pr
	updatedPR.AssignedReviewers = []string{"u3"}

	oldUser := models.User{UserID: "u2", TeamName: "backend"}
	team := models.Team{
		Name: "backend",
		Members: []models.TeamMember{
			{UserID: "u2", IsActive: true},
			{UserID: "u3", IsActive: true},
		},
	}

	prRepo.On("GetPR", mock.Anything, "pr-1001").Return(pr, nil).Once()
	userRepo.On("GetUser", mock.Anything, "u2").Return(oldUser, nil)
	teamRepo.On("GetTeam", mock.Anything, "backend").Return(team, nil)
	prRepo.On("ReassignReviewer", mock.Anything, "pr-1001", "u2", "u3").Return(nil)
	prRepo.On("GetPR", mock.Anything, "pr-1001").Return(updatedPR, nil).Once()

	uc := NewPRUsecase(prRepo, userRepo, teamRepo, testLogger())
	req := models.ReassignRequest{PRID: "pr-1001", OldReviewerID: "u2"}
	newPR, replacedBy, err := uc.ReassignReviewer(context.Background(), req)
	require.NoError(t, err)
	require.Equal(t, "u3", replacedBy)
	require.Contains(t, newPR.AssignedReviewers, "u3")

	prRepo.AssertExpectations(t)
	userRepo.AssertExpectations(t)
	teamRepo.AssertExpectations(t)
}

func TestPRUsecase_ReassignReviewer_NoCandidate(t *testing.T) {
	prRepo := new(mockPRRepository)
	userRepo := new(mockUserRepository)
	teamRepo := new(mockTeamRepository)

	pr := models.PullRequest{ID: "pr-1001", Status: models.StatusOpen, AssignedReviewers: []string{"u2"}}
	oldUser := models.User{UserID: "u2", TeamName: "backend"}
	team := models.Team{Name: "backend", Members: []models.TeamMember{{UserID: "u2", IsActive: true}}}

	prRepo.On("GetPR", mock.Anything, "pr-1001").Return(pr, nil)
	userRepo.On("GetUser", mock.Anything, "u2").Return(oldUser, nil)
	teamRepo.On("GetTeam", mock.Anything, "backend").Return(team, nil)

	uc := NewPRUsecase(prRepo, userRepo, teamRepo, testLogger())
	_, _, err := uc.ReassignReviewer(context.Background(), models.ReassignRequest{PRID: "pr-1001", OldReviewerID: "u2"})
	require.ErrorIs(t, err, models.ErrNoCandidate)

	prRepo.AssertExpectations(t)
	userRepo.AssertExpectations(t)
	teamRepo.AssertExpectations(t)
}

func TestPRUsecase_GetPRsByReviewer(t *testing.T) {
	prRepo := new(mockPRRepository)
	userRepo := new(mockUserRepository)
	teamRepo := new(mockTeamRepository)

	prs := []models.PullRequest{
		{ID: "pr-1001", Name: "Fix", AuthorID: "u1", Status: models.StatusOpen},
	}

	prRepo.On("GetPRsByReviewer", mock.Anything, "u2").Return(prs, nil)

	uc := NewPRUsecase(prRepo, userRepo, teamRepo, testLogger())

	result, err := uc.GetPRsByReviewer(context.Background(), "u2")

	require.NoError(t, err)
	require.Len(t, result, 1)
	require.Equal(t, "pr-1001", result[0].ID)

	prRepo.AssertExpectations(t)
	userRepo.AssertExpectations(t)
	teamRepo.AssertExpectations(t)
}

func TestPRUsecase_ReassignReviewer_AuthorNotCandidate(t *testing.T) {
	prRepo := new(mockPRRepository)
	userRepo := new(mockUserRepository)
	teamRepo := new(mockTeamRepository)

	pr := models.PullRequest{
		ID:                "pr-1001",
		Status:            models.StatusOpen,
		AuthorID:          "u1",
		AssignedReviewers: []string{"u2"},
	}

	updatedPR := pr
	updatedPR.AssignedReviewers = []string{"u3"}

	oldUser := models.User{UserID: "u2", TeamName: "backend"}
	team := models.Team{
		Name: "backend",
		Members: []models.TeamMember{
			{UserID: "u1", IsActive: true},
			{UserID: "u2", IsActive: true},
			{UserID: "u3", IsActive: true},
		},
	}

	prRepo.On("GetPR", mock.Anything, "pr-1001").Return(pr, nil).Once()
	userRepo.On("GetUser", mock.Anything, "u2").Return(oldUser, nil)
	teamRepo.On("GetTeam", mock.Anything, "backend").Return(team, nil)
	prRepo.On("ReassignReviewer", mock.Anything, "pr-1001", "u2", "u3").Return(nil)
	prRepo.On("GetPR", mock.Anything, "pr-1001").Return(updatedPR, nil).Once()

	uc := NewPRUsecase(prRepo, userRepo, teamRepo, testLogger())
	req := models.ReassignRequest{PRID: "pr-1001", OldReviewerID: "u2"}
	_, replacedBy, err := uc.ReassignReviewer(context.Background(), req)
	require.NoError(t, err)
	require.Equal(t, "u3", replacedBy)
	require.NotEqual(t, "u1", replacedBy)

	prRepo.AssertExpectations(t)
	userRepo.AssertExpectations(t)
	teamRepo.AssertExpectations(t)
}
func TestPRUsecase_GetUserStats_Success(t *testing.T) {
	prRepo := new(mockPRRepository)
	userRepo := new(mockUserRepository)
	teamRepo := new(mockTeamRepository)

	stats := []models.UserStats{
		{
			UserID:          "u1",
			TeamName:        "backend",
			Username:        "Alice",
			AssignmentCount: 3,
			AssignedPRs:     []string{"pr-1", "pr-2", "pr-3"},
		},
	}

	prRepo.On("GetUserStats", mock.Anything).Return(stats, nil)

	uc := NewPRUsecase(prRepo, userRepo, teamRepo, testLogger())

	result, err := uc.GetUserStats(context.Background())

	require.NoError(t, err)
	require.Len(t, result, 1)
	require.Equal(t, 3, result[0].AssignmentCount)
	require.Equal(t, []string{"pr-1", "pr-2", "pr-3"}, result[0].AssignedPRs)

	prRepo.AssertExpectations(t)
	userRepo.AssertExpectations(t)
	teamRepo.AssertExpectations(t)
}
