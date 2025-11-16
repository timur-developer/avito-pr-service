package postgres

import (
	"avito-pr-service/internal/models"
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func insertUserTestData(t *testing.T, db *pgxpool.Pool) {
	ctx := context.Background()
	_, err := db.Exec(ctx, `
		INSERT INTO teams (name) VALUES ('team1');
		INSERT INTO users (user_id, username, team_name, is_active) VALUES 
			('u1', 'User1', 'team1', true),
			('u2', 'User2', 'team1', false);
	`)
	require.NoError(t, err)
}

func TestUserRepository_Integration_GetUser(t *testing.T) {
	dbPool, cleanup := setupTestDB(t)
	defer cleanup()
	insertUserTestData(t, dbPool)

	repo := newUserRepository(dbPool)

	ctx := context.Background()
	u, err := repo.GetUser(ctx, "u1")
	require.NoError(t, err)
	assert.Equal(t, "u1", u.UserID)
	assert.Equal(t, "User1", u.Username)
	assert.True(t, u.IsActive)

	_, err = repo.GetUser(ctx, "nonexistent")
	assert.Equal(t, models.ErrUserNotFound, err)
}

func TestUserRepository_Integration_SetActive(t *testing.T) {
	dbPool, cleanup := setupTestDB(t)
	defer cleanup()
	insertUserTestData(t, dbPool)

	repo := newUserRepository(dbPool)

	ctx := context.Background()
	err := repo.SetActive(ctx, "u1", false)
	require.NoError(t, err)

	u, err := repo.GetUser(ctx, "u1")
	require.NoError(t, err)
	assert.False(t, u.IsActive)

	err = repo.SetActive(ctx, "nonexistent", true)
	assert.Equal(t, models.ErrUserNotFound, err)
}

func TestUserRepository_Integration_DeactivateTeam(t *testing.T) {
	dbPool, cleanup := setupTestDB(t)
	defer cleanup()
	insertUserTestData(t, dbPool)

	repo := newUserRepository(dbPool)

	ctx := context.Background()
	count, err := repo.DeactivateTeam(ctx, "team1")
	require.NoError(t, err)
	assert.Equal(t, 1, count)

	u1, err := repo.GetUser(ctx, "u1")
	require.NoError(t, err)
	assert.False(t, u1.IsActive)

	u2, err := repo.GetUser(ctx, "u2")
	require.NoError(t, err)
	assert.False(t, u2.IsActive)

	count, err = repo.DeactivateTeam(ctx, "nonexistent")
	require.NoError(t, err)
	assert.Equal(t, 0, count)
}
