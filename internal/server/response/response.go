package response

import (
	"avito-pr-service/internal/models"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-playground/validator/v10"
	"net/http"
	"strings"
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
		case models.ErrorTeamExists, models.ErrorPRExists, models.ErrorPRMerged, models.ErrorNoCandidate, models.ErrorNotAssigned:
			status = http.StatusConflict
		case models.ErrorNotFound:
			status = http.StatusNotFound
		}
		JSON(w, map[string]any{
			"error": map[string]string{
				"code":    string(appErr.Code),
				"message": appErr.Message,
			},
		}, status)
		return
	}

	JSON(w, map[string]any{
		"error": map[string]string{
			"code":    "INTERNAL",
			"message": "server error",
		},
	}, http.StatusInternalServerError)
}

func ValidationError(w http.ResponseWriter, err error) {
	var validationErrors validator.ValidationErrors
	if !errors.As(err, &validationErrors) {
		BadRequest(w, "invalid request")
		return
	}

	messages := make([]string, 0, len(validationErrors))
	for _, e := range validationErrors {
		field := e.Field()
		switch e.Tag() {
		case "required":
			messages = append(messages, fmt.Sprintf("%s is required", field))
		case "min":
			messages = append(messages, fmt.Sprintf("%s must be at least %s characters", field, e.Param()))
		default:
			messages = append(messages, fmt.Sprintf("%s is invalid", field))
		}
	}

	JSON(w, map[string]any{
		"error": map[string]string{
			"code":    "VALIDATION_ERROR",
			"message": strings.Join(messages, ", "),
		},
	}, http.StatusBadRequest)
}

func BadRequest(w http.ResponseWriter, message string) {
	JSON(w, map[string]any{
		"error": map[string]string{
			"code":    "VALIDATION_ERROR",
			"message": message,
		},
	}, http.StatusBadRequest)
}
