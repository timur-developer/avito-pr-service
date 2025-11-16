package repository

import (
	"avito-pr-service/internal/models"
	"context"
	"time"
)

type TeamRepository interface {
	CreateTeam(ctx context.Context, team models.Team) error
	GetTeam(ctx context.Context, name string) (models.Team, error)
}

type UserRepository interface {
	SetActive(ctx context.Context, userID string, isActive bool) error
	GetUser(ctx context.Context, userID string) (models.User, error)
	DeactivateTeam(ctx context.Context, teamName string) (int, error)
}

type PRRepository interface {
	CreatePR(ctx context.Context, pr models.PullRequest) error
	GetPR(ctx context.Context, prID string) (models.PullRequest, error)
	MergePR(ctx context.Context, prID string) (*time.Time, error)
	ReassignReviewer(ctx context.Context, prID, oldUID, newUID string) error
	GetPRsByReviewer(ctx context.Context, userID string) ([]models.PullRequest, error)
	GetUserStats(ctx context.Context) ([]models.UserStats, error)
	GetOpenPRsWithTeamReviewers(ctx context.Context, teamName string) ([]models.PullRequest, error)
}
