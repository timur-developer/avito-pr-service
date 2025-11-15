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
	r.Post("/pullRequest/merge", h.MergePR)
	r.Post("/pullRequest/reassign", h.ReassignReviewer)
	r.Get("/users/getReview", h.GetPRsByReviewer)
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

func (h *PRHandler) MergePR(w http.ResponseWriter, r *http.Request) {
	var req models.MergePRRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid JSON")
		return
	}
	if err := models.Validate(&req); err != nil {
		response.ValidationError(w, err)
		return
	}

	pr, err := h.uc.MergePR(r.Context(), req.PRID)
	if err != nil {
		response.Error(w, err, http.StatusInternalServerError)
		return
	}
	response.JSON(w, map[string]any{"pr": pr}, http.StatusOK)
}

func (h *PRHandler) ReassignReviewer(w http.ResponseWriter, r *http.Request) {
	var req models.ReassignRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid JSON")
		return
	}
	if err := models.Validate(&req); err != nil {
		response.ValidationError(w, err)
		return
	}

	pr, replacedBy, err := h.uc.ReassignReviewer(r.Context(), models.ReassignRequest{
		PRID:          req.PRID,
		OldReviewerID: req.OldReviewerID,
	})
	if err != nil {
		response.Error(w, err, http.StatusInternalServerError)
		return
	}
	response.JSON(w, map[string]any{
		"pr":          pr,
		"replaced_by": replacedBy,
	}, http.StatusOK)
}

func (h *PRHandler) GetPRsByReviewer(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		response.BadRequest(w, "user_id is required")
		return
	}
	prs, err := h.uc.GetPRsByReviewer(r.Context(), userID)
	if err != nil {
		response.Error(w, err, http.StatusInternalServerError)
		return
	}
	response.JSON(w, map[string]any{"pull_requests": prs}, http.StatusOK)
}
