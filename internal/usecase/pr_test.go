package usecase

import (
	"context"
	"testing"

	"avito-pr-service/internal/models"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type mockPRRepository struct{ mock.Mock }

func (m *mockPRRepository) CreatePR(ctx context.Context, pr models.PullRequest) error {
	return m.Called(ctx, pr).Error(0)
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
		return pr.ID == "pr-1001" &&
			pr.Name == "Add search" &&
			len(pr.AssignedReviewers) == 2
	})).Return(nil)

	uc := NewPRUsecase(prRepo, userRepo, teamRepo, testLogger())

	req := models.CreatePRRequest{
		ID:       "pr-1001",
		Name:     "Add search",
		AuthorID: "u1",
	}

	pr, err := uc.CreatePR(context.Background(), req)

	require.NoError(t, err)
	require.Equal(t, "pr-1001", pr.ID)
	require.Len(t, pr.AssignedReviewers, 2)

	prRepo.AssertExpectations(t)
	userRepo.AssertExpectations(t)
	teamRepo.AssertExpectations(t)
}
