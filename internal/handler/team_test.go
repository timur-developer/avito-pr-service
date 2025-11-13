package handler

import (
	"avito-pr-service/internal/models"
	"bytes"
	"context"
	"encoding/json"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/require"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
)

type mockTeamUsecase struct{}

func (m *mockTeamUsecase) AddTeam(ctx context.Context, team models.Team) error {
	if team.Name == "exists" {
		return models.ErrTeamExists
	}
	return nil
}

func (m *mockTeamUsecase) GetTeam(ctx context.Context, name string) (models.Team, error) {
	if name == "avito" {
		return models.Team{
			Name: "avito",
			Members: []models.TeamMember{
				{UserID: "u1", Username: "alice", IsActive: true},
			},
		}, nil
	}
	return models.Team{}, models.ErrNotFound
}

func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func TestTeamHandler_AddTeam_Success(t *testing.T) {
	uc := &mockTeamUsecase{}
	h := NewTeamHandler(uc, testLogger())

	r := chi.NewRouter()
	h.Register(r)

	body := `{"team_name":"new-team","members":[{"user_id":"u1","username":"alice","is_active":true}]}`
	req := httptest.NewRequest(http.MethodPost, "/team/add", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	var resp map[string]string
	json.Unmarshal(w.Body.Bytes(), &resp)
	require.Equal(t, "ok", resp["status"])
}

func TestTeamHandler_AddTeam_Exists(t *testing.T) {
	uc := &mockTeamUsecase{}
	h := NewTeamHandler(uc, testLogger())

	r := chi.NewRouter()
	h.Register(r)

	body := `{"team_name":"exists","members":[{"user_id":"u1","username":"alice","is_active":true}]}`
	req := httptest.NewRequest(http.MethodPost, "/team/add", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusConflict, w.Code)
	var resp map[string]any
	json.Unmarshal(w.Body.Bytes(), &resp)
	err := resp["error"].(map[string]any)
	require.Equal(t, "TEAM_EXISTS", err["code"])
}

func TestTeamHandler_GetTeam_Success(t *testing.T) {
	uc := &mockTeamUsecase{}
	h := NewTeamHandler(uc, testLogger())

	r := chi.NewRouter()
	h.Register(r)

	req := httptest.NewRequest(http.MethodGet, "/team/get?team_name=avito", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	var resp models.Team
	json.Unmarshal(w.Body.Bytes(), &resp)
	require.Equal(t, "avito", resp.Name)
	require.Len(t, resp.Members, 1)
}

func TestTeamHandler_GetTeam_NotFound(t *testing.T) {
	uc := &mockTeamUsecase{}
	h := NewTeamHandler(uc, testLogger())

	r := chi.NewRouter()
	h.Register(r)

	req := httptest.NewRequest(http.MethodGet, "/team/get?team_name=unknown", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusNotFound, w.Code)
}

func TestTeamHandler_AddTeam_Validation_Fail(t *testing.T) {
	uc := &mockTeamUsecase{}
	h := NewTeamHandler(uc, testLogger())

	r := chi.NewRouter()
	h.Register(r)

	body := `{"team_name":""}`
	req := httptest.NewRequest(http.MethodPost, "/team/add", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusBadRequest, w.Code)
	var resp map[string]any
	json.Unmarshal(w.Body.Bytes(), &resp)
	err := resp["error"].(map[string]any)
	require.Contains(t, err["message"].(string), "Name")
}

func TestTeamHandler_GetTeam_MissingParam(t *testing.T) {
	uc := &mockTeamUsecase{}
	h := NewTeamHandler(uc, testLogger())

	r := chi.NewRouter()
	h.Register(r)

	req := httptest.NewRequest(http.MethodGet, "/team/get", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusBadRequest, w.Code)
	var resp map[string]any
	json.Unmarshal(w.Body.Bytes(), &resp)
	err := resp["error"].(map[string]any)
	require.Contains(t, err["message"].(string), "team_name required")
}
