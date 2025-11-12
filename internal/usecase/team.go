package usecase

import (
	"avito-pr-service/internal/models"
	"avito-pr-service/internal/repository"
	"context"
	"errors"
)

type TeamUsecase interface {
	AddTeam(ctx context.Context, team models.Team) error
	GetTeam(ctx context.Context, name string) (models.Team, error)
}

type teamUsecase struct {
	repo repository.TeamRepository
}

func NewTeamUsecase(repo repository.TeamRepository) TeamUsecase {
	return &teamUsecase{repo: repo}
}

func (u *teamUsecase) AddTeam(ctx context.Context, team models.Team) error {
	_, err := u.repo.GetTeam(ctx, team.Name)
	if err == nil {
		return models.ErrTeamExists
	}
	if !errors.Is(err, repository.ErrNotFound) {
		return err
	}
	return u.repo.CreateTeam(ctx, team)
}

func (u *teamUsecase) GetTeam(ctx context.Context, name string) (models.Team, error) {
	team, err := u.repo.GetTeam(ctx, name)
	if errors.Is(err, repository.ErrNotFound) {
		return team, models.ErrNotFound
	}
	return team, err
}
