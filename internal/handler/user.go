package handler

import (
	"avito-pr-service/internal/models"
	"avito-pr-service/internal/server/response"
	"avito-pr-service/internal/usecase"
	"encoding/json"
	"github.com/go-chi/chi/v5"
	"log/slog"
	"net/http"
)

type UserHandler struct {
	uc  usecase.UserUsecase
	log *slog.Logger
}

func NewUserHandler(uc usecase.UserUsecase, log *slog.Logger) *UserHandler {
	return &UserHandler{
		uc:  uc,
		log: log.With("handler", "user"),
	}
}

func (h *UserHandler) Register(r chi.Router) {
	r.Post("/users/setIsActive", h.SetActive)
}

func (h *UserHandler) SetActive(w http.ResponseWriter, r *http.Request) {
	h.log.Info("SetActive request")

	var req models.SetUserActiveRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid JSON")
		return
	}

	if err := models.Validate(&req); err != nil {
		response.ValidationError(w, err)
		return
	}

	if err := h.uc.SetActive(r.Context(), req); err != nil {
		response.Error(w, err, http.StatusInternalServerError)
		return
	}

	user, err := h.uc.GetUser(r.Context(), req.UserID)
	if err != nil {
		response.Error(w, err, http.StatusInternalServerError)
		return
	}

	response.JSON(w, user, http.StatusOK)
}
