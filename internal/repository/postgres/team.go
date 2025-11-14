package postgres

import (
	"avito-pr-service/internal/models"
	"avito-pr-service/internal/repository"
	"context"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type teamRepository struct {
	db *pgxpool.Pool
}

func newTeamRepository(db *pgxpool.Pool) repository.TeamRepository {
	return &teamRepository{db: db}
}

func (r *teamRepository) CreateTeam(ctx context.Context, team models.Team) error {
	tx, err := r.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO teams (name) VALUES ($1) ON CONFLICT DO NOTHING
	`, team.Name)
	if err != nil {
		tx.Rollback(ctx)
		return fmt.Errorf("insert team: %w", err)
	}

	for _, m := range team.Members {
		_, err = tx.Exec(ctx, `
			INSERT INTO users (user_id, username, team_name)
    VALUES ($1, $2, $3)
    ON CONFLICT (user_id) DO UPDATE SET
        username = EXCLUDED.username,
        team_name = EXCLUDED.team_name
		`, m.UserID, m.Username, team.Name)
		if err != nil {
			tx.Rollback(ctx)
			return fmt.Errorf("upsert user %s: %w", m.UserID, err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		tx.Rollback(ctx)
		return fmt.Errorf("commit transaction: %w", err)
	}
	return nil
}

func (r *teamRepository) GetTeam(ctx context.Context, name string) (models.Team, error) {
	var team models.Team
	team.Name = name

	rows, err := r.db.Query(ctx, `
		SELECT user_id, username, is_active
		FROM users
		WHERE team_name = $1
	`, name)
	if err != nil {
		return team, fmt.Errorf("query: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var m models.TeamMember
		if err := rows.Scan(&m.UserID, &m.Username, &m.IsActive); err != nil {
			return team, fmt.Errorf("scan: %w", err)
		}
		team.Members = append(team.Members, m)
	}

	if err := rows.Err(); err != nil {
		return team, fmt.Errorf("rows: %w", err)
	}

	if len(team.Members) == 0 {
		return team, models.ErrTeamNotFound
	}

	return team, nil
}
