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
	DeactivateTeam(ctx context.Context, req models.DeactivateTeamRequest) (models.DeactivateTeamResponse, error)
}

type teamUsecase struct {
	repo     repository.TeamRepository
	userRepo repository.UserRepository
	prRepo   repository.PRRepository
	prUC     PRUsecase
	log      *slog.Logger
}

func NewTeamUsecase(repo repository.TeamRepository, userRepo repository.UserRepository, prRepo repository.PRRepository, prUC PRUsecase, log *slog.Logger) TeamUsecase {
	return &teamUsecase{
		repo:     repo,
		userRepo: userRepo,
		prRepo:   prRepo,
		prUC:     prUC,
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

func (u *teamUsecase) DeactivateTeam(ctx context.Context, req models.DeactivateTeamRequest) (models.DeactivateTeamResponse, error) {
	resp := models.DeactivateTeamResponse{}

	prs, err := u.prRepo.GetOpenPRsWithTeamReviewers(ctx, req.TeamName)
	if err != nil {
		u.log.Error("failed to get PRs", "error", err)
		return resp, err
	}

	for _, pr := range prs {
		for _, reviewerID := range pr.AssignedReviewers {
			user, err := u.userRepo.GetUser(ctx, reviewerID)
			if err != nil || user.TeamName != req.TeamName {
				continue
			}

			reassignReq := models.ReassignRequest{
				PRID:          pr.ID,
				OldReviewerID: reviewerID,
			}

			_, _, err = u.prUC.ReassignReviewer(ctx, reassignReq)
			if err == nil {
				resp.ReassignedPRs++
			} else if !errors.Is(err, models.ErrNoCandidate) {
				u.log.Warn("failed to reassign", "pr", pr.ID, "old", reviewerID, "error", err)
			}
		}
	}

	deactivated, err := u.userRepo.DeactivateTeam(ctx, req.TeamName)
	if err != nil {
		u.log.Error("failed to deactivate team", "team", req.TeamName, "error", err)
		return resp, err
	}
	resp.DeactivatedUsers = deactivated

	u.log.Info("team deactivated", "team", req.TeamName, "users", deactivated, "reassigned", resp.ReassignedPRs)
	return resp, nil
}
