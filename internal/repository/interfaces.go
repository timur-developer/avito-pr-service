package repository

import (
	"avito-pr-service/internal/models"
	"context"
)

type TeamRepository interface {
	CreateTeam(ctx context.Context, team models.Team) error
	GetTeam(ctx context.Context, name string) (models.Team, error)
}

type UserRepository interface {
	SetActive(ctx context.Context, userID string, isActive bool) error
	GetUser(ctx context.Context, userID string) (models.User, error)
}

type PRRepository interface {
	CreatePR(ctx context.Context, pr models.PullRequest) error
}
