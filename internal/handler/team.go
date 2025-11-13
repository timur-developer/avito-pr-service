package handler

import (
	"avito-pr-service/internal/models"
	"avito-pr-service/internal/server"
	"avito-pr-service/internal/usecase"
	"encoding/json"
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
		server.BadRequest(w, "invalid JSON")
		return
	}

	if err := models.Validate(&team); err != nil {
		h.log.Warn("validation failed", "error", err, "team", team)
		server.BadRequest(w, err.Error())
		return
	}

	if err := h.uc.AddTeam(r.Context(), team); err != nil {
		h.log.Error("usecase error", "error", err, "team_name", team.Name)
		server.Error(w, err, http.StatusInternalServerError)
		return
	}

	h.log.Info("team created", "team_name", team.Name, "members_count", len(team.Members))
	server.JSON(w, map[string]string{"status": "ok"}, http.StatusOK)
}

func (h *TeamHandler) GetTeam(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("team_name")
	if name == "" {
		server.BadRequest(w, "team_name required")
		return
	}

	team, err := h.uc.GetTeam(r.Context(), name)
	if err != nil {
		server.Error(w, err, http.StatusInternalServerError)
		return
	}

	server.JSON(w, team, http.StatusOK)
}
