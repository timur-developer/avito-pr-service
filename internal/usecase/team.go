package usecase

import (
	"avito-pr-service/internal/models"
	"avito-pr-service/internal/repository"
	"context"
	"errors"
	"log/slog"
)

type TeamUsecase interface {
	AddTeam(ctx context.Context, team models.Team) error
	GetTeam(ctx context.Context, name string) (models.Team, error)
}

type teamUsecase struct {
	repo     repository.TeamRepository
	userRepo repository.UserRepository
	log      *slog.Logger
}

func NewTeamUsecase(repo repository.TeamRepository, userRepo repository.UserRepository, log *slog.Logger) TeamUsecase {
	return &teamUsecase{
		repo:     repo,
		userRepo: userRepo,
		log:      log.With("layer", "usecase", "entity", "team"),
	}
}

func (u *teamUsecase) AddTeam(ctx context.Context, team models.Team) error {
	seen := make(map[string]bool)
	for _, m := range team.Members {
		if seen[m.UserID] {
			return models.ErrDuplicateUserID
		}
		seen[m.UserID] = true
	}

	for _, m := range team.Members {
		user, err := u.userRepo.GetUser(ctx, m.UserID)
		if err == nil && user.TeamName != "" && user.TeamName != team.Name {
			return models.ErrUserInAnotherTeam
		}
		if err != nil && !errors.Is(err, models.ErrUserNotFound) {
			return err
		}
	}

	_, err := u.repo.GetTeam(ctx, team.Name)
	if err == nil {
		u.log.Warn("team already exists", "team_name", team.Name)
		return models.ErrTeamExists
	}
	if !errors.Is(err, models.ErrTeamNotFound) {
		u.log.Error("database error on GetTeam", "error", err)
		return err
	}

	u.log.Info("creating new team", "team_name", team.Name)
	return u.repo.CreateTeam(ctx, team)
}

func (u *teamUsecase) GetTeam(ctx context.Context, name string) (models.Team, error) {
	team, err := u.repo.GetTeam(ctx, name)
	if errors.Is(err, models.ErrTeamNotFound) {
		return team, models.ErrTeamNotFound
	}
	return team, err
}
