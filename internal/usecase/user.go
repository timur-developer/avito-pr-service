package usecase

import (
	"avito-pr-service/internal/models"
	"avito-pr-service/internal/repository"
	"context"
	"errors"
	"log/slog"
)

type UserUsecase interface {
	SetActive(ctx context.Context, req models.SetUserActiveRequest) error
	GetUser(ctx context.Context, userID string) (models.User, error)
}

type userUsecase struct {
	repo repository.UserRepository
	log  *slog.Logger
}

func NewUserUsecase(repo repository.UserRepository, log *slog.Logger) UserUsecase {
	return &userUsecase{
		repo: repo,
		log:  log.With("layer", "usecase", "entity", "user"),
	}
}

func (u *userUsecase) SetActive(ctx context.Context, req models.SetUserActiveRequest) error {
	u.log.Info("setting user active", "user_id", req.UserID, "is_active", req.IsActive)

	if err := u.repo.SetActive(ctx, req.UserID, req.IsActive); err != nil {
		if errors.Is(err, models.ErrUserNotFound) {
			u.log.Warn("user not found", "user_id", req.UserID)
			return models.ErrUserNotFound
		}
		u.log.Error("failed to update user", "error", err)
		return err
	}

	u.log.Info("user updated", "user_id", req.UserID, "is_active", req.IsActive)
	return nil
}

func (u *userUsecase) GetUser(ctx context.Context, userID string) (models.User, error) {
	return u.repo.GetUser(ctx, userID)
}
