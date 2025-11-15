package postgres

import (
	"avito-pr-service/internal/models"
	"avito-pr-service/internal/repository"
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type prRepository struct {
	db *pgxpool.Pool
}

func newPrRepository(db *pgxpool.Pool) repository.PRRepository {
	return &prRepository{db: db}
}

func (r *prRepository) CreatePR(ctx context.Context, pr models.PullRequest) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("beign tx: %w", err)
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, `
        INSERT INTO pull_requests (id, name, author_id, status, created_at)
        VALUES ($1, $2, $3, 'OPEN', $4)
    `, pr.ID, pr.Name, pr.AuthorID, pr.CreatedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.ConstraintName == "pull_requests_pkey" {
			return models.ErrPRExists
		}
		return fmt.Errorf("insert pr: %w", err)
	}

	for _, uid := range pr.AssignedReviewers {
		_, err = tx.Exec(ctx, `INSERT INTO pr_reviewers (pr_id, user_id) VALUES ($1, $2)`, pr.ID, uid)
		if err != nil {
			return fmt.Errorf("insert reviewer: %w", err)
		}
	}

	return tx.Commit(ctx)
}

func (r *prRepository) GetPR(ctx context.Context, prID string) (models.PullRequest, error) {
	var pr models.PullRequest

	err := r.db.QueryRow(ctx, `
        SELECT id, name, author_id, status, created_at, merged_at
        FROM pull_requests WHERE id = $1
    `, prID).Scan(&pr.ID, &pr.Name, &pr.AuthorID, &pr.Status, &pr.CreatedAt, &pr.MergedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.PullRequest{}, models.ErrNotFound
		}
		return models.PullRequest{}, err
	}

	rows, err := r.db.Query(ctx, `SELECT user_id FROM pr_reviewers WHERE pr_id = $1`, prID)
	if err != nil {
		return models.PullRequest{}, err
	}
	defer rows.Close()

	for rows.Next() {
		var uid string
		if err := rows.Scan(&uid); err != nil {
			return models.PullRequest{}, err
		}
		pr.AssignedReviewers = append(pr.AssignedReviewers, uid)
	}
	return pr, nil
}

func (r *prRepository) MergePR(ctx context.Context, prID string) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	var status string
	err = tx.QueryRow(ctx, `SELECT status FROM pull_requests WHERE id = $1 FOR UPDATE`, prID).
		Scan(&status)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.ErrNotFound
		}
		return err
	}
	if status == "MERGED" {
		return models.ErrPRMerged
	}
	if status != "OPEN" {
		return models.ErrInvalidStatus
	}

	_, err = tx.Exec(ctx, `
        UPDATE pull_requests 
        SET status = 'MERGED', merged_at = NOW() 
        WHERE id = $1
    `, prID)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (r *prRepository) ReassignReviewer(ctx context.Context, prID, oldUID, newUID string) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, `DELETE FROM pr_reviewers WHERE pr_id = $1 AND user_id = $2`, prID, oldUID)
	if err != nil {
		return err
	}

	_, err = tx.Exec(ctx, `INSERT INTO pr_reviewers (pr_id, user_id) VALUES ($1, $2)`, prID, newUID)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return models.ErrAlreadyAssigned
		}
		return err
	}

	return tx.Commit(ctx)
}

func (r *prRepository) GetPRsByReviewer(ctx context.Context, userID string) ([]models.PullRequest, error) {
	prIDsQuery := `
        SELECT DISTINCT pr.pr_id 
        FROM pr_reviewers pr 
        WHERE pr.user_id = $1
    `

	rows, err := r.db.Query(ctx, prIDsQuery, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var prIDs []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		prIDs = append(prIDs, id)
	}

	if len(prIDs) == 0 {
		return []models.PullRequest{}, nil
	}

	prs := make([]models.PullRequest, 0, len(prIDs))
	for _, prID := range prIDs {
		pr, err := r.GetPR(ctx, prID)
		if err != nil {
			return nil, err
		}
		prs = append(prs, pr)
	}

	return prs, nil
}

func (r *prRepository) GetUserStats(ctx context.Context) ([]models.UserStats, error) {
	rows, err := r.db.Query(ctx, `
        SELECT 
            u.user_id, 
            u.team_name, 
            u.username, 
            COUNT(pr.pr_id) as count,
            COALESCE(array_agg(pr.pr_id) FILTER (WHERE pr.pr_id IS NOT NULL), '{}') as pr_ids
        FROM users u
        LEFT JOIN pr_reviewers pr ON u.user_id = pr.user_id
        GROUP BY u.user_id
        ORDER BY count DESC
    `)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stats []models.UserStats
	for rows.Next() {
		var s models.UserStats
		var prIDs []string
		if err := rows.Scan(&s.UserID, &s.TeamName, &s.Username, &s.AssignmentCount, &prIDs); err != nil {
			return nil, err
		}
		s.AssignedPRs = prIDs
		stats = append(stats, s)
	}
	return stats, nil
}
