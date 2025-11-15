package postgres

import (
	"avito-pr-service/internal/models"
	"avito-pr-service/internal/repository"
	"context"
	"errors"
	"fmt"
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
