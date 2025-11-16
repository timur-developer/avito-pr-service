// internal/repository/postgres/pr_integration_test.go
package postgres

import (
	"avito-pr-service/internal/models"
	"context"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func insertPRTestData(t *testing.T, db *pgxpool.Pool) {
	ctx := context.Background()
	_, err := db.Exec(ctx, `
		INSERT INTO teams (name) VALUES ('team1');
		INSERT INTO users (user_id, username, team_name, is_active) VALUES 
			('u1', 'Author', 'team1', true),
			('u2', 'Reviewer1', 'team1', true),
			('u3', 'Reviewer2', 'team1', true),
			('u4', 'Reviewer3', 'team1', true);
	`)
	require.NoError(t, err, "failed to insert test data")
}

func TestPRRepository_Integration_CreatePR(t *testing.T) {
	dbPool, cleanup := setupTestDB(t)
	defer cleanup()
	insertPRTestData(t, dbPool)

	repo := newPrRepository(dbPool)

	ctx := context.Background()
	now := time.Now()
	pr := models.PullRequest{
		ID:                "pr-1",
		Name:              "Test PR",
		AuthorID:          "u1",
		AssignedReviewers: []string{"u2", "u3"},
		CreatedAt:         &now,
	}

	err := repo.CreatePR(ctx, pr)
	require.NoError(t, err, "failed to create PR")

	gotPR, err := repo.GetPR(ctx, "pr-1")
	require.NoError(t, err)
	assert.Equal(t, "pr-1", gotPR.ID)
	assert.Equal(t, "OPEN", gotPR.Status)
	assert.ElementsMatch(t, []string{"u2", "u3"}, gotPR.AssignedReviewers)
}

func TestPRRepository_Integration_GetPR(t *testing.T) {
	dbPool, cleanup := setupTestDB(t)
	defer cleanup()
	insertPRTestData(t, dbPool)

	repo := newPrRepository(dbPool)

	ctx := context.Background()
	now := time.Now()
	pr := models.PullRequest{
		ID:                "pr-1",
		Name:              "Test PR",
		AuthorID:          "u1",
		AssignedReviewers: []string{"u2"},
		CreatedAt:         &now,
	}
	err := repo.CreatePR(ctx, pr)
	require.NoError(t, err)

	gotPR, err := repo.GetPR(ctx, "pr-1")
	require.NoError(t, err)
	assert.Equal(t, "Test PR", gotPR.Name)
	assert.Len(t, gotPR.AssignedReviewers, 1)
}

func TestPRRepository_Integration_MergePR(t *testing.T) {
	dbPool, cleanup := setupTestDB(t)
	defer cleanup()
	insertPRTestData(t, dbPool)

	repo := newPrRepository(dbPool)

	ctx := context.Background()
	now := time.Now()
	pr := models.PullRequest{
		ID:                "pr-1",
		Name:              "Test PR",
		AuthorID:          "u1",
		AssignedReviewers: []string{"u2"},
		CreatedAt:         &now,
	}
	err := repo.CreatePR(ctx, pr)
	require.NoError(t, err)

	mergedAt, err := repo.MergePR(ctx, "pr-1")
	require.NoError(t, err)
	assert.NotNil(t, mergedAt)

	gotPR, err := repo.GetPR(ctx, "pr-1")
	require.NoError(t, err)
	assert.Equal(t, models.StatusMerged, gotPR.Status)
	assert.NotNil(t, gotPR.MergedAt)

	_, err = repo.MergePR(ctx, "pr-1")
	assert.Equal(t, models.ErrPRMerged, err)
}

func TestPRRepository_Integration_ReassignReviewer(t *testing.T) {
	dbPool, cleanup := setupTestDB(t)
	defer cleanup()
	insertPRTestData(t, dbPool)

	repo := newPrRepository(dbPool)

	ctx := context.Background()
	now := time.Now()
	pr := models.PullRequest{
		ID:                "pr-1",
		Name:              "Test PR",
		AuthorID:          "u1",
		AssignedReviewers: []string{"u2", "u3"},
		CreatedAt:         &now,
	}
	err := repo.CreatePR(ctx, pr)
	require.NoError(t, err)

	err = repo.ReassignReviewer(ctx, "pr-1", "u2", "u4")
	require.NoError(t, err)

	gotPR, err := repo.GetPR(ctx, "pr-1")
	require.NoError(t, err)
	assert.ElementsMatch(t, []string{"u3", "u4"}, gotPR.AssignedReviewers)

	_, err = repo.MergePR(ctx, "pr-1")
	require.NoError(t, err)
	err = repo.ReassignReviewer(ctx, "pr-1", "u3", "u4")
	assert.Equal(t, models.ErrPRMerged, err)
}

func TestPRRepository_Integration_GetPRsByReviewer(t *testing.T) {
	dbPool, cleanup := setupTestDB(t)
	defer cleanup()
	insertPRTestData(t, dbPool)

	repo := newPrRepository(dbPool)

	ctx := context.Background()
	now1 := time.Now()
	pr1 := models.PullRequest{
		ID:                "pr-1",
		Name:              "Test PR1",
		AuthorID:          "u1",
		AssignedReviewers: []string{"u2"},
		CreatedAt:         &now1,
	}
	err := repo.CreatePR(ctx, pr1)
	require.NoError(t, err)

	now2 := time.Now()
	pr2 := models.PullRequest{
		ID:                "pr-2",
		Name:              "Test PR2",
		AuthorID:          "u1",
		AssignedReviewers: []string{"u2", "u3"},
		CreatedAt:         &now2,
	}
	err = repo.CreatePR(ctx, pr2)
	require.NoError(t, err)

	prs, err := repo.GetPRsByReviewer(ctx, "u2")
	require.NoError(t, err)
	assert.Len(t, prs, 2)
	assert.Equal(t, "pr-1", prs[0].ID)
	assert.Equal(t, "pr-2", prs[1].ID)
}

func TestPRRepository_Integration_GetUserStats(t *testing.T) {
	dbPool, cleanup := setupTestDB(t)
	defer cleanup()
	insertPRTestData(t, dbPool)

	repo := newPrRepository(dbPool)

	ctx := context.Background()
	now := time.Now()
	pr := models.PullRequest{
		ID:                "pr-1",
		Name:              "Test PR",
		AuthorID:          "u1",
		AssignedReviewers: []string{"u2", "u3"},
		CreatedAt:         &now,
	}
	err := repo.CreatePR(ctx, pr)
	require.NoError(t, err)

	stats, err := repo.GetUserStats(ctx)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(stats), 2)

	var u2Stat, u3Stat models.UserStats
	for _, s := range stats {
		if s.UserID == "u2" {
			u2Stat = s
		} else if s.UserID == "u3" {
			u3Stat = s
		}
	}
	assert.Equal(t, 1, u2Stat.AssignmentCount)
	assert.ElementsMatch(t, []string{"pr-1"}, u2Stat.AssignedPRs)
	assert.Equal(t, 1, u3Stat.AssignmentCount)
}

func TestPRRepository_Integration_GetOpenPRsWithTeamReviewers(t *testing.T) {
	dbPool, cleanup := setupTestDB(t)
	defer cleanup()
	insertPRTestData(t, dbPool)

	repo := newPrRepository(dbPool)

	ctx := context.Background()
	now := time.Now()
	pr := models.PullRequest{
		ID:                "pr-1",
		Name:              "Test PR",
		AuthorID:          "u1",
		AssignedReviewers: []string{"u2", "u3"},
		CreatedAt:         &now,
	}
	err := repo.CreatePR(ctx, pr)
	require.NoError(t, err)

	prs, err := repo.GetOpenPRsWithTeamReviewers(ctx, "team1")
	require.NoError(t, err)
	assert.Len(t, prs, 1)
	assert.Equal(t, "pr-1", prs[0].ID)
	assert.Equal(t, models.StatusOpen, prs[0].Status)
	assert.ElementsMatch(t, []string{"u2", "u3"}, prs[0].AssignedReviewers)

	_, err = repo.MergePR(ctx, "pr-1")
	require.NoError(t, err)
	prsAfter, err := repo.GetOpenPRsWithTeamReviewers(ctx, "team1")
	require.NoError(t, err)
	assert.Len(t, prsAfter, 0)
}
