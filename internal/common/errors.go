package common

import (
	"errors"
	"net/http"
)

type AppError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Status  int    `json:"status"`
	Details map[string]string `json:"fields,omitempty"`
}

func (e *AppError) Error() string {
	return e.Message
}

func NewAppError(code, message string, status int) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
		Status:  status,
	}
}

var (
	ErrNotFound           = NewAppError("NOT_FOUND", "Resource not found", http.StatusNotFound)
	ErrUserAlreadyExists  = NewAppError("USER_ALREADY_EXISTS", "User already exists", http.StatusBadRequest)
	ErrInvalidLoginCredentials = NewAppError("INVALID_CREDENTIALS", "Invalid email or password", http.StatusUnauthorized)
	ErrInvalidVerificationCredential         = NewAppError("INVALID_VERIFICATION_CREDENTIAL", "Invalid verification credential", http.StatusBadRequest)
	ErrUnauthorized       = NewAppError("UNAUTHORIZED", "Unauthorized access", http.StatusUnauthorized)
	ErrInternal           = NewAppError("INTERNAL_ERROR", "Something went wrong", http.StatusInternalServerError)
	ErrValidation		 = NewAppError("VALIDATION_ERROR", "Validation failed", http.StatusUnprocessableEntity)
	ErrBadRequest		 = NewAppError("BAD_REQUEST", "Bad request", http.StatusBadRequest)
	ErrUserNotVerified   = NewAppError("USER_NOT_VERIFIED", "User email/phone not verified", http.StatusForbidden)
)

func IsAppError(err error) (*AppError, bool) {
	var appErr *AppError
	ok := errors.As(err, &appErr)
	return appErr, ok
}