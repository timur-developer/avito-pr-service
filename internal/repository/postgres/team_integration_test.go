package postgres

import (
	"avito-pr-service/internal/models"
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestTeamRepository_Integration_CreateTeam(t *testing.T) {
	dbPool, cleanup := setupTestDB(t)
	defer cleanup()

	repo := newTeamRepository(dbPool)

	ctx := context.Background()
	team := models.Team{
		Name: "team1",
		Members: []models.TeamMember{
			{UserID: "u1", Username: "User1", IsActive: true},
			{UserID: "u2", Username: "User2", IsActive: false},
		},
	}
	err := repo.CreateTeam(ctx, team)
	require.NoError(t, err)
	gotTeam, err := repo.GetTeam(ctx, "team1")
	require.NoError(t, err)
	assert.Equal(t, "team1", gotTeam.Name)
	assert.Len(t, gotTeam.Members, 2)
	assert.Equal(t, "u1", gotTeam.Members[0].UserID)
	assert.True(t, gotTeam.Members[0].IsActive)

	team.Members[0].IsActive = false
	err = repo.CreateTeam(ctx, team)
	require.NoError(t, err)

	gotTeam, err = repo.GetTeam(ctx, "team1")
	require.NoError(t, err)
	assert.False(t, gotTeam.Members[0].IsActive)
}

func TestTeamRepository_Integration_GetTeam(t *testing.T) {
	dbPool, cleanup := setupTestDB(t)
	defer cleanup()

	repo := newTeamRepository(dbPool)

	ctx := context.Background()
	team := models.Team{
		Name: "team1",
		Members: []models.TeamMember{
			{UserID: "u1", Username: "User1", IsActive: true},
		},
	}
	err := repo.CreateTeam(ctx, team)
	require.NoError(t, err)

	gotTeam, err := repo.GetTeam(ctx, "team1")
	require.NoError(t, err)
	assert.Equal(t, "team1", gotTeam.Name)
	assert.Len(t, gotTeam.Members, 1)

	_, err = repo.GetTeam(ctx, "nonexistent")
	assert.Equal(t, models.ErrTeamNotFound, err)
}
