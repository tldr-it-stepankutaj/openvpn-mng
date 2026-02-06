package apperror

import "net/http"

// AppError represents a structured application error with HTTP status code
type AppError struct {
	Code    int    `json:"code"`
	Err     string `json:"error"`
	Message string `json:"message"`
}

func (e *AppError) Error() string {
	return e.Message
}

func NotFound(msg string) *AppError {
	return &AppError{Code: http.StatusNotFound, Err: "Not Found", Message: msg}
}

func Validation(msg string) *AppError {
	return &AppError{Code: http.StatusBadRequest, Err: "Bad Request", Message: msg}
}

func Conflict(msg string) *AppError {
	return &AppError{Code: http.StatusConflict, Err: "Conflict", Message: msg}
}

func Forbidden(msg string) *AppError {
	return &AppError{Code: http.StatusForbidden, Err: "Forbidden", Message: msg}
}

func Unauthorized(msg string) *AppError {
	return &AppError{Code: http.StatusUnauthorized, Err: "Unauthorized", Message: msg}
}

func TooManyRequests(msg string) *AppError {
	return &AppError{Code: http.StatusTooManyRequests, Err: "Too Many Requests", Message: msg}
}

func Internal(msg string) *AppError {
	return &AppError{Code: http.StatusInternalServerError, Err: "Internal Server Error", Message: msg}
}
