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

type DeactivateTeamRequest struct {
	TeamName string `json:"team_name" validate:"required"`
}

type DeactivateTeamResponse struct {
	DeactivatedUsers int `json:"deactivated_users"`
	ReassignedPRs    int `json:"reassigned_prs"`
}
