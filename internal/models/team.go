package models

type Team struct {
	Name    string       `json:"team_name" validate:"required"`
	Members []TeamMember `json:"members" validate:"required,dive"`
}

type TeamMember struct {
	UserID   string `json:"user_id" validate:"required"`
	Username string `json:"username" validate:"required"`
	IsActive bool   `json:"is_active"`
}
