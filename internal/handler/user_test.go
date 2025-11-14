// internal/handler/user_test.go
package handler

import (
	"avito-pr-service/internal/models"
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/require"
)

type mockUserUsecase struct {
	users map[string]models.User
}

func (m *mockUserUsecase) SetActive(ctx context.Context, req models.SetUserActiveRequest) error {
	user, exists := m.users[req.UserID]
	if !exists {
		return models.ErrUserNotFound
	}
	user.IsActive = req.IsActive
	m.users[req.UserID] = user
	return nil
}

func (m *mockUserUsecase) GetUser(ctx context.Context, userID string) (models.User, error) {
	user, exists := m.users[userID]
	if !exists {
		return models.User{}, models.ErrUserNotFound
	}
	return user, nil
}

func TestUserHandler_SetActive_Success_ToFalse(t *testing.T) {
	uc := &mockUserUsecase{
		users: map[string]models.User{
			"u1": {UserID: "u1", Username: "alice", TeamName: "alpha", IsActive: true},
		},
	}
	h := NewUserHandler(uc, testLogger())
	r := chi.NewRouter()
	h.Register(r)

	body := `{"user_id":"u1","is_active":false}`
	req := httptest.NewRequest(http.MethodPost, "/users/setIsActive", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	var resp models.User
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	require.Equal(t, "u1", resp.UserID)
	require.False(t, resp.IsActive)
	require.Equal(t, "alice", resp.Username)
}

func TestUserHandler_SetActive_Success_ToTrue(t *testing.T) {
	uc := &mockUserUsecase{
		users: map[string]models.User{
			"u2": {UserID: "u2", Username: "bob", TeamName: "beta", IsActive: false},
		},
	}
	h := NewUserHandler(uc, testLogger())
	r := chi.NewRouter()
	h.Register(r)

	body := `{"user_id":"u2","is_active":true}`
	req := httptest.NewRequest(http.MethodPost, "/users/setIsActive", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	var resp models.User
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	require.True(t, resp.IsActive)
}

func TestUserHandler_SetActive_NotFound(t *testing.T) {
	uc := &mockUserUsecase{users: make(map[string]models.User)}
	h := NewUserHandler(uc, testLogger())
	r := chi.NewRouter()
	h.Register(r)

	body := `{"user_id":"u404","is_active":false}`
	req := httptest.NewRequest(http.MethodPost, "/users/setIsActive", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusNotFound, w.Code)

	var resp map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	errMap := resp["error"].(map[string]any)
	require.Equal(t, "NOT_FOUND", errMap["code"])
}

func TestUserHandler_SetActive_Validation_NoUserID(t *testing.T) {
	uc := &mockUserUsecase{users: make(map[string]models.User)}
	h := NewUserHandler(uc, testLogger())
	r := chi.NewRouter()
	h.Register(r)

	body := `{"is_active":true}`
	req := httptest.NewRequest(http.MethodPost, "/users/setIsActive", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusBadRequest, w.Code)

	var resp map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	errMap := resp["error"].(map[string]any)
	require.Contains(t, errMap["message"].(string), "user_id is required")
}

func TestUserHandler_SetActive_InvalidJSON(t *testing.T) {
	uc := &mockUserUsecase{users: make(map[string]models.User)}
	h := NewUserHandler(uc, testLogger())
	r := chi.NewRouter()
	h.Register(r)

	body := `{"user_id":"u1", "is_active":}`
	req := httptest.NewRequest(http.MethodPost, "/users/setIsActive", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusBadRequest, w.Code)

	var resp map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	errMap := resp["error"].(map[string]any)
	require.Equal(t, "VALIDATION_ERROR", errMap["code"])
	require.Contains(t, errMap["message"].(string), "invalid JSON")
}
