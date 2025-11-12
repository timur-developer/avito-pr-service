package models

type ErrorCode string

const (
	ErrorTeamExists  ErrorCode = "TEAM_EXISTS"
	ErrorPRExists    ErrorCode = "PR_EXISTS"
	ErrorPRMerged    ErrorCode = "PR_MERGED"
	ErrorNotAssigned ErrorCode = "NOT_ASSIGNED"
	ErrorNoCandidate ErrorCode = "NO_CANDIDATE"
	ErrorNotFound    ErrorCode = "NOT_FOUND"
)

type AppError struct {
	Code    ErrorCode `json:"code"`
	Message string    `json:"message"`
}

func (e AppError) Error() string {
	return e.Message
}

var (
	ErrTeamExists  = AppError{Code: ErrorTeamExists, Message: "team already exists"}
	ErrNotFound    = AppError{Code: ErrorNotFound, Message: "team not found"}
	ErrPRExists    = AppError{Code: ErrorPRExists, Message: "pull request already exists"}
	ErrNotAssigned = AppError{Code: ErrorNotAssigned, Message: "not assigned"}
	ErrNoCandidate = AppError{Code: ErrorNoCandidate, Message: "no candidate"}
	ErrPRMerged    = AppError{Code: ErrorPRMerged, Message: "pull request already merged"}
)
