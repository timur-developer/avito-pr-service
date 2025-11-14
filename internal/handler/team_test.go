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

type mockTeamUsecase struct {
	teams map[string]models.Team
}

func (m *mockTeamUsecase) AddTeam(ctx context.Context, team models.Team) error {
	if team.Name == "exists" {
		return models.ErrTeamExists
	}
	
	for i := range team.Members {
		team.Members[i].IsActive = true
	}

	m.teams[team.Name] = team
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
	team, ok := m.teams[name]
	if !ok {
		return models.Team{}, models.ErrTeamNotFound
	}
	return team, nil
}

func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func TestTeamHandler_AddTeam_Success(t *testing.T) {
	uc := &mockTeamUsecase{teams: make(map[string]models.Team)}
	h := NewTeamHandler(uc, testLogger())
	r := chi.NewRouter()
	h.Register(r)

	body := `{"team_name":"new-team","members":[{"user_id":"u1","username":"alice"}]}`
	req := httptest.NewRequest(http.MethodPost, "/team/add", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	var resp models.Team
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	require.Equal(t, "new-team", resp.Name)
	require.Len(t, resp.Members, 1)
	require.True(t, resp.Members[0].IsActive)
}

func TestTeamHandler_AddTeam_Exists(t *testing.T) {
	uc := &mockTeamUsecase{teams: make(map[string]models.Team)}
	h := NewTeamHandler(uc, testLogger())
	r := chi.NewRouter()
	h.Register(r)

	body := `{"team_name":"exists","members":[{"user_id":"u1","username":"alice"}]}`
	req := httptest.NewRequest(http.MethodPost, "/team/add", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusConflict, w.Code)

	var resp map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	errMap := resp["error"].(map[string]any)
	require.Equal(t, "TEAM_EXISTS", errMap["code"])
}

func TestTeamHandler_AddTeam_Validation_Fail(t *testing.T) {
	uc := &mockTeamUsecase{teams: make(map[string]models.Team)}
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
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	errMap := resp["error"].(map[string]any)
	require.Contains(t, errMap["message"].(string), "name is required")
	require.Contains(t, errMap["message"].(string), "members is required")
}
