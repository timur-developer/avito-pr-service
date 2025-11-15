package usecase

import (
	"avito-pr-service/internal/models"
	"avito-pr-service/internal/repository"
	"avito-pr-service/internal/utils"
	"context"
	"errors"
	"log/slog"
	"time"
)

type PRUsecase interface {
	CreatePR(ctx context.Context, req models.CreatePRRequest) (models.PullRequest, error)
}

type prUsecase struct {
	prRepo   repository.PRRepository
	userRepo repository.UserRepository
	teamRepo repository.TeamRepository
	log      *slog.Logger
}

func NewPRUsecase(pr repository.PRRepository, user repository.UserRepository, team repository.TeamRepository, log *slog.Logger) PRUsecase {
	return &prUsecase{pr, user, team, log}
}

func (u *prUsecase) CreatePR(ctx context.Context, req models.CreatePRRequest) (models.PullRequest, error) {
	u.log.Info("creating PR", "id", req.ID, "author", req.AuthorID)

	author, err := u.userRepo.GetUser(ctx, req.AuthorID)
	if err != nil || !author.IsActive {
		if errors.Is(err, models.ErrUserNotFound) {
			return models.PullRequest{}, models.ErrUserNotFound
		}
		return models.PullRequest{}, models.ErrNotFound
	}

	team, err := u.teamRepo.GetTeam(ctx, author.TeamName)
	if err != nil {
		return models.PullRequest{}, models.ErrNotFound
	}

	candidates := []string{}
	for _, m := range team.Members {
		if m.UserID != req.AuthorID && m.IsActive {
			candidates = append(candidates, m.UserID)
		}
	}

	reviewers := utils.PickRandom(candidates, 2)
	if reviewers == nil {
		reviewers = []string{}
	}

	pr := models.PullRequest{
		ID:                req.ID,
		Name:              req.Name,
		AuthorID:          req.AuthorID,
		Status:            "OPEN",
		AssignedReviewers: reviewers,
		CreatedAt:         utils.Ptr(time.Now()),
	}

	if err := u.prRepo.CreatePR(ctx, pr); err != nil {
		return models.PullRequest{}, err
	}

	return pr, nil
}
