package usecase

import (
	"avito-pr-service/internal/models"
	"avito-pr-service/internal/repository"
	"avito-pr-service/internal/utils"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"slices"
	"time"
)

type PRUsecase interface {
	CreatePR(ctx context.Context, req models.CreatePRRequest) (models.PullRequest, error)
	GetPR(ctx context.Context, prID string) (models.PullRequest, error)
	MergePR(ctx context.Context, prID string) (models.PullRequest, error)
	ReassignReviewer(ctx context.Context, req models.ReassignRequest) (models.PullRequest, string, error)
	GetPRsByReviewer(ctx context.Context, userID string) ([]models.PullRequest, error)
	GetUserStats(ctx context.Context) ([]models.UserStats, error)
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
	if err != nil {
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
		Status:            models.StatusOpen,
		AssignedReviewers: reviewers,
		CreatedAt:         utils.Ptr(time.Now()),
	}

	if err := u.prRepo.CreatePR(ctx, pr); err != nil {
		return models.PullRequest{}, err
	}

	return pr, nil
}

func (u *prUsecase) GetPR(ctx context.Context, prID string) (models.PullRequest, error) {
	return u.prRepo.GetPR(ctx, prID)
}

func (u *prUsecase) MergePR(ctx context.Context, prID string) (models.PullRequest, error) {
	pr, err := u.prRepo.GetPR(ctx, prID)
	if err != nil {
		if errors.Is(err, models.ErrNotFound) {
			return models.PullRequest{}, models.ErrPRNotFound
		}
		return models.PullRequest{}, err
	}

	if pr.Status == models.StatusMerged {
		return pr, nil
	}
	if pr.Status != models.StatusOpen {
		return models.PullRequest{}, models.ErrInvalidStatus
	}

	mergedAt, err := u.prRepo.MergePR(ctx, prID)
	if err != nil {
		return models.PullRequest{}, err
	}

	pr.Status = models.StatusMerged
	pr.MergedAt = mergedAt
	return pr, nil
}

func (u *prUsecase) ReassignReviewer(ctx context.Context, req models.ReassignRequest) (models.PullRequest, string, error) {
	pr, err := u.prRepo.GetPR(ctx, req.PRID)
	if err != nil {
		return models.PullRequest{}, "", err
	}

	if pr.Status != models.StatusOpen {
		return models.PullRequest{}, "", models.ErrPRMerged
	}
	if !slices.Contains(pr.AssignedReviewers, req.OldReviewerID) {
		return models.PullRequest{}, "", models.ErrNotAssigned
	}

	oldUser, err := u.userRepo.GetUser(ctx, req.OldReviewerID)
	if err != nil {
		return models.PullRequest{}, "", err
	}

	team, err := u.teamRepo.GetTeam(ctx, oldUser.TeamName)
	if err != nil {
		return models.PullRequest{}, "", err
	}

	candidates := []string{}
	for _, m := range team.Members {
		if m.UserID == pr.AuthorID {
			continue
		}
		if m.IsActive && !slices.Contains(pr.AssignedReviewers, m.UserID) {
			candidates = append(candidates, m.UserID)
		}
	}

	if len(candidates) == 0 {
		return models.PullRequest{}, "", models.ErrNoCandidate
	}

	newUID := utils.PickRandom(candidates, 1)[0]

	if err := u.prRepo.ReassignReviewer(ctx, req.PRID, req.OldReviewerID, newUID); err != nil {
		return models.PullRequest{}, "", err
	}

	freshPR, err := u.prRepo.GetPR(ctx, req.PRID)
	if err != nil {
		return models.PullRequest{}, "", err
	}

	return freshPR, newUID, nil
}

func (u *prUsecase) GetPRsByReviewer(ctx context.Context, userID string) ([]models.PullRequest, error) {
	_, err := u.userRepo.GetUser(ctx, userID)
	if err != nil {
		if errors.Is(err, models.ErrNotFound) {
			return nil, models.ErrUserNotFound
		}
		return nil, fmt.Errorf("get user: %w", err)
	}

	return u.prRepo.GetPRsByReviewer(ctx, userID)
}

func (u *prUsecase) GetUserStats(ctx context.Context) ([]models.UserStats, error) {
	return u.prRepo.GetUserStats(ctx)
}
