package handler

import (
	"avito-pr-service/internal/utils"
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"avito-pr-service/internal/models"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/require"
)

type mockPRUsecase struct {
	prs map[string]models.PullRequest
}

func (m *mockPRUsecase) CreatePR(ctx context.Context, req models.CreatePRRequest) (models.PullRequest, error) {
	if req.ID == "exists" {
		return models.PullRequest{}, models.ErrPRExists
	}
	pr := models.PullRequest{
		ID:                req.ID,
		Name:              req.Name,
		AuthorID:          req.AuthorID,
		Status:            "OPEN",
		AssignedReviewers: []string{"u2", "u3"},
		CreatedAt:         utils.Ptr(time.Now()),
	}
	m.prs[req.ID] = pr
	return pr, nil
}

func TestPRHandler_CreatePR_Success(t *testing.T) {
	uc := &mockPRUsecase{prs: make(map[string]models.PullRequest)}
	h := NewPRHandler(uc, testLogger())
	r := chi.NewRouter()
	h.Register(r)

	body := `{"pull_request_id":"pr-1001","pull_request_name":"Add search","author_id":"u1"}`
	req := httptest.NewRequest(http.MethodPost, "/pullRequest/create", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusCreated, w.Code)

	var resp struct {
		PR models.PullRequest `json:"pr"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))

	require.Equal(t, "pr-1001", resp.PR.ID)
	require.Len(t, resp.PR.AssignedReviewers, 2)
	require.Equal(t, "OPEN", resp.PR.Status)
}

func TestPRHandler_CreatePR_Exists(t *testing.T) {
	uc := &mockPRUsecase{prs: map[string]models.PullRequest{"exists": {}}}
	h := NewPRHandler(uc, testLogger())
	r := chi.NewRouter()
	h.Register(r)

	body := `{"pull_request_id":"exists","pull_request_name":"test","author_id":"u1"}`
	req := httptest.NewRequest(http.MethodPost, "/pullRequest/create", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusConflict, w.Code)

	var resp map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	errMap := resp["error"].(map[string]any)
	require.Equal(t, "PR_EXISTS", errMap["code"])
}
