package handler

import (
	"avito-pr-service/internal/models"
	"avito-pr-service/internal/server/response"
	"avito-pr-service/internal/usecase"
	"encoding/json"
	"errors"
	"github.com/go-chi/chi/v5"
	"log/slog"
	"net/http"
)

type TeamHandler struct {
	uc  usecase.TeamUsecase
	log *slog.Logger
}

func NewTeamHandler(uc usecase.TeamUsecase, log *slog.Logger) *TeamHandler {
	return &TeamHandler{
		uc:  uc,
		log: log.With("handler", "team")}
}

func (h *TeamHandler) Register(r chi.Router) {
	r.Post("/team/add", h.AddTeam)
	r.Get("/team/get", h.GetTeam)
}

func (h *TeamHandler) AddTeam(w http.ResponseWriter, r *http.Request) {
	h.log.Info("AddTeam request", "remote_addr", r.RemoteAddr, "user_agent", r.UserAgent())

	var team models.Team
	if err := json.NewDecoder(r.Body).Decode(&team); err != nil {
		h.log.Warn("invalid JSON", "error", err)
		response.BadRequest(w, "invalid JSON")
		return
	}

	if err := models.Validate(&team); err != nil {
		h.log.Warn("validation failed", "error", err, "team", team)
		response.ValidationError(w, err)
		return
	}

	if err := h.uc.AddTeam(r.Context(), team); err != nil {
		h.log.Error("usecase error", "error", err, "team_name", team.Name)
		if errors.Is(err, models.ErrUserInAnotherTeam) {
			response.Error(w, models.ErrUserInAnotherTeam, http.StatusBadRequest)
			return
		}
		response.Error(w, err, http.StatusInternalServerError)
		return
	}

	createdTeam, err := h.uc.GetTeam(r.Context(), team.Name)
	if err != nil {
		response.Error(w, err, http.StatusInternalServerError)
	}

	h.log.Info("team created", "team_name", team.Name, "members_count", len(team.Members))
	response.JSON(w, createdTeam, http.StatusCreated)
}

func (h *TeamHandler) GetTeam(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("team_name")
	if name == "" {
		response.BadRequest(w, "team_name required")
		return
	}

	team, err := h.uc.GetTeam(r.Context(), name)
	if err != nil {
		response.Error(w, err, http.StatusInternalServerError)
		return
	}

	response.JSON(w, team, http.StatusOK)
}
