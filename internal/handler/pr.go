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

type PRHandler struct {
	uc  usecase.PRUsecase
	log *slog.Logger
}

func NewPRHandler(uc usecase.PRUsecase, log *slog.Logger) *PRHandler {
	return &PRHandler{
		uc:  uc,
		log: log.With("handler", "pull_request")}
}

func (h *PRHandler) Register(r chi.Router) {
	r.Post("/pullRequest/create", h.CreatePR)
}

func (h *PRHandler) CreatePR(w http.ResponseWriter, r *http.Request) {
	var req models.CreatePRRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid JSON")
		return
	}
	if err := models.Validate(&req); err != nil {
		response.ValidationError(w, err)
		return
	}

	pr, err := h.uc.CreatePR(r.Context(), req)
	if err != nil {
		response.Error(w, err, http.StatusConflict)
		return
	}

	response.JSON(w, map[string]any{"pr": pr}, http.StatusCreated)
}
