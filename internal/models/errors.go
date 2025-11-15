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

func (e AppError) Is(target error) bool {
	t, ok := target.(AppError)
	if !ok {
		return false
	}
	return e.Code == t.Code
}

var (
	ErrTeamExists   = AppError{Code: ErrorTeamExists, Message: "team already exists"}
	ErrTeamNotFound = AppError{Code: ErrorNotFound, Message: "team not found"}
	ErrUserNotFound = AppError{Code: ErrorNotFound, Message: "user not found"}
	ErrNotFound     = AppError{Code: ErrorNotFound, Message: "not found"}
	ErrPRExists     = AppError{Code: ErrorPRExists, Message: "pull request already exists"}
	ErrNotAssigned  = AppError{Code: ErrorNotAssigned, Message: "not assigned"}
	ErrNoCandidate  = AppError{Code: ErrorNoCandidate, Message: "no candidate"}
	ErrPRMerged     = AppError{Code: ErrorPRMerged, Message: "pull request already merged"}
)
