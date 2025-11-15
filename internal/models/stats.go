package models

type UserStats struct {
	UserID          string   `json:"user_id"`
	TeamName        string   `json:"team_name"`
	Username        string   `json:"username"`
	AssignmentCount int      `json:"assignment_count"`
	AssignedPRs     []string `json:"assigned_prs"`
}
