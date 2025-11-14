package models

type User struct {
	UserID   string `json:"user_id" validate:"required"`
	Username string `json:"username" validate:"min=1"`
	TeamName string `json:"team_name" validate:"required"`
	IsActive bool   `json:"is_active"`
}

type SetUserActiveRequest struct {
	UserID   string `json:"user_id" validate:"required"`
	IsActive bool   `json:"is_active"`
}
