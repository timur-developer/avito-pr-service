// internal/handler/pr_test.go
package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"avito-pr-service/internal/models"
	"avito-pr-service/internal/utils"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type mockPRUsecase struct{ mock.Mock }

func (m *mockPRUsecase) CreatePR(ctx context.Context, req models.CreatePRRequest) (models.PullRequest, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(models.PullRequest), args.Error(1)
}
func (m *mockPRUsecase) MergePR(ctx context.Context, prID string) (models.PullRequest, error) {
	args := m.Called(ctx, prID)
	return args.Get(0).(models.PullRequest), args.Error(1)
}
func (m *mockPRUsecase) ReassignReviewer(ctx context.Context, req models.ReassignRequest) (models.PullRequest, string, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(models.PullRequest), args.String(1), args.Error(2)
}
func (m *mockPRUsecase) GetPRsByReviewer(ctx context.Context, userID string) ([]models.PullRequest, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]models.PullRequest), args.Error(1)
}

func (m *mockPRUsecase) GetPR(ctx context.Context, prID string) (models.PullRequest, error) {
	args := m.Called(ctx, prID)
	return args.Get(0).(models.PullRequest), args.Error(1)
}
func TestPRHandler_CreatePR_Success(t *testing.T) {
	uc := new(mockPRUsecase)
	h := NewPRHandler(uc, testLogger())
	r := chi.NewRouter()
	h.Register(r)

	pr := models.PullRequest{
		ID: "pr-1001", Name: "Add search", AuthorID: "u1", Status: "OPEN",
		AssignedReviewers: []string{"u2", "u3"}, CreatedAt: utils.Ptr(time.Now()),
	}
	uc.On("CreatePR", mock.Anything, mock.Anything).Return(pr, nil)

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
}

func TestPRHandler_CreatePR_Exists(t *testing.T) {
	uc := new(mockPRUsecase)
	h := NewPRHandler(uc, testLogger())
	r := chi.NewRouter()
	h.Register(r)

	uc.On("CreatePR", mock.Anything, mock.Anything).Return(models.PullRequest{}, models.ErrPRExists)

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

func TestPRHandler_MergePR_Success(t *testing.T) {
	uc := new(mockPRUsecase)
	h := NewPRHandler(uc, testLogger())
	r := chi.NewRouter()
	h.Register(r)

	mergedPR := models.PullRequest{
		ID: "pr-1001", Status: "MERGED", MergedAt: utils.Ptr(time.Now()),
	}
	uc.On("MergePR", mock.Anything, "pr-1001").Return(mergedPR, nil)

	body := `{"pull_request_id":"pr-1001"}`
	req := httptest.NewRequest(http.MethodPost, "/pullRequest/merge", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	var resp struct {
		PR models.PullRequest `json:"pr"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	require.Equal(t, "MERGED", resp.PR.Status)
}

func TestPRHandler_MergePR_NotFound(t *testing.T) {
	uc := new(mockPRUsecase)
	h := NewPRHandler(uc, testLogger())
	r := chi.NewRouter()
	h.Register(r)

	uc.On("MergePR", mock.Anything, "pr-404").Return(models.PullRequest{}, models.ErrNotFound)

	body := `{"pull_request_id":"pr-404"}`
	req := httptest.NewRequest(http.MethodPost, "/pullRequest/merge", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusNotFound, w.Code)
}

func TestPRHandler_ReassignReviewer_Success(t *testing.T) {
	uc := new(mockPRUsecase)
	h := NewPRHandler(uc, testLogger())
	r := chi.NewRouter()
	h.Register(r)

	pr := models.PullRequest{ID: "pr-1001", AssignedReviewers: []string{"u3"}}
	uc.On("ReassignReviewer", mock.Anything, mock.Anything).Return(pr, "u3", nil)

	body := `{"pull_request_id":"pr-1001","old_reviewer_id":"u2"}`
	req := httptest.NewRequest(http.MethodPost, "/pullRequest/reassign", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	var resp struct {
		PR         models.PullRequest `json:"pr"`
		ReplacedBy string             `json:"replaced_by"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	require.Equal(t, "u3", resp.ReplacedBy)
}

func TestPRHandler_GetPRsByReviewer_Success(t *testing.T) {
	uc := new(mockPRUsecase)
	h := NewPRHandler(uc, testLogger())
	r := chi.NewRouter()
	h.Register(r)

	prs := []models.PullRequest{{ID: "pr-1001", Status: "OPEN"}}
	uc.On("GetPRsByReviewer", mock.Anything, "u2").Return(prs, nil)

	req := httptest.NewRequest(http.MethodGet, "/users/getReview?user_id=u2", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	var resp struct {
		PullRequests []models.PullRequest `json:"pull_requests"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	require.Len(t, resp.PullRequests, 1)
}
