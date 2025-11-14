package postgres

import (
	"avito-pr-service/internal/models"
	"avito-pr-service/internal/repository"
	"context"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type userRepository struct {
	db *pgxpool.Pool
}

func newUserRepository(db *pgxpool.Pool) repository.UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) SetActive(ctx context.Context, userID string, isActive bool) error {
	result, err := r.db.Exec(ctx, `
        UPDATE users 
        SET is_active = $1 
        WHERE user_id = $2
    `, isActive, userID)
	if err != nil {
		return fmt.Errorf("update user: %w", err)
	}

	if result.RowsAffected() == 0 {
		return models.ErrUserNotFound
	}

	return nil
}

func (r *userRepository) GetUser(ctx context.Context, userID string) (models.User, error) {
	var u models.User
	err := r.db.QueryRow(ctx, `
        SELECT user_id, username, team_name, is_active 
        FROM users WHERE user_id = $1
    `, userID).Scan(&u.UserID, &u.Username, &u.TeamName, &u.IsActive)
	if err != nil {
		if err == pgx.ErrNoRows {
			return u, models.ErrUserNotFound
		}
		return u, fmt.Errorf("query user: %w", err)
	}
	return u, nil
}
