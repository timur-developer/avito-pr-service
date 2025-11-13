package server

import (
	"avito-pr-service/internal/models"
	"encoding/json"
	"errors"
	"net/http"
)

func JSON(w http.ResponseWriter, data any, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func Error(w http.ResponseWriter, err error, defaultStatus int) {
	var appErr models.AppError
	if errors.As(err, &appErr) {
		status := defaultStatus
		switch appErr.Code {
		case models.ErrorTeamExists, models.ErrorPRExists:
			status = http.StatusConflict
		case models.ErrorNotFound:
			status = http.StatusNotFound
		case models.ErrorNoCandidate:
			status = http.StatusNotFound
		case models.ErrorNotAssigned:
			status = http.StatusBadRequest
		}
		JSON(w, map[string]any{
			"error": map[string]string{
				"code":    string(appErr.Code),
				"message": appErr.Message,
			},
		}, status)
	} else {
		JSON(w, map[string]any{
			"error": map[string]string{
				"code":    "INTERNAL",
				"message": "server error",
			},
		}, http.StatusInternalServerError)
	}
}

func BadRequest(w http.ResponseWriter, message string) {
	JSON(w, map[string]any{
		"error": map[string]string{
			"code":    "VALIDATION_ERROR",
			"message": message,
		},
	}, http.StatusBadRequest)
}
