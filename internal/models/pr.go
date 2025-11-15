package models

import "time"

type PullRequest struct {
	ID                string     `json:"pull_request_id"`
	Name              string     `json:"pull_request_name"`
	AuthorID          string     `json:"author_id"`
	Status            string     `json:"status"`
	AssignedReviewers []string   `json:"assigned_reviewers"`
	CreatedAt         *time.Time `json:"createdAt,omitempty"`
	MergedAt          *time.Time `json:"mergedAt,omitempty"`
}

type CreatePRRequest struct {
	ID       string `json:"pull_request_id" validate:"required"`
	Name     string `json:"pull_request_name" validate:"required,min=1"`
	AuthorID string `json:"author_id" validate:"required"`
}
