package errors

import (
	"fmt"
	"net/http"
)

// AppError represents application-specific errors
type AppError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Status  int    `json:"-"`
}

func (e *AppError) Error() string {
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// Predefined error codes
const (
	// Validation errors
	ErrCodeValidation = "VALIDATION_ERROR"
	ErrCodeRequired   = "REQUIRED_FIELD"
	ErrCodeInvalid    = "INVALID_VALUE"

	// Business logic errors
	ErrCodeNotFound         = "NOT_FOUND"
	ErrCodeAlreadyExists    = "ALREADY_EXISTS"
	ErrCodeInsufficientFund = "INSUFFICIENT_FUND"
	ErrCodeInvalidSplit     = "INVALID_SPLIT"
	ErrCodeCurrencyMismatch = "CURRENCY_MISMATCH"

	// System errors
	ErrCodeDatabase    = "DATABASE_ERROR"
	ErrCodeInternal    = "INTERNAL_ERROR"
	ErrCodeIdempotency = "IDEMPOTENCY_ERROR"
)

// Validation errors
func NewValidationError(message string) *AppError {
	return &AppError{
		Code:    ErrCodeValidation,
		Message: message,
		Status:  http.StatusBadRequest,
	}
}

func NewRequiredFieldError(field string) *AppError {
	return &AppError{
		Code:    ErrCodeRequired,
		Message: fmt.Sprintf("Field '%s' is required", field),
		Status:  http.StatusBadRequest,
	}
}

func NewInvalidValueError(field, value string) *AppError {
	return &AppError{
		Code:    ErrCodeInvalid,
		Message: fmt.Sprintf("Invalid value '%s' for field '%s'", value, field),
		Status:  http.StatusBadRequest,
	}
}

// Business logic errors
func NewNotFoundError(resource string) *AppError {
	return &AppError{
		Code:    ErrCodeNotFound,
		Message: fmt.Sprintf("%s not found", resource),
		Status:  http.StatusNotFound,
	}
}

func NewAlreadyExistsError(resource string) *AppError {
	return &AppError{
		Code:    ErrCodeAlreadyExists,
		Message: fmt.Sprintf("%s already exists", resource),
		Status:  http.StatusConflict,
	}
}

func NewInsufficientFundError(available, required string) *AppError {
	return &AppError{
		Code:    ErrCodeInsufficientFund,
		Message: fmt.Sprintf("Insufficient funds: available %s, required %s", available, required),
		Status:  http.StatusBadRequest,
	}
}

func NewInvalidSplitError(message string) *AppError {
	return &AppError{
		Code:    ErrCodeInvalidSplit,
		Message: message,
		Status:  http.StatusBadRequest,
	}
}

func NewCurrencyMismatchError() *AppError {
	return &AppError{
		Code:    ErrCodeCurrencyMismatch,
		Message: "Currency mismatch in transaction",
		Status:  http.StatusBadRequest,
	}
}

// System errors
func NewDatabaseError(err error) *AppError {
	return &AppError{
		Code:    ErrCodeDatabase,
		Message: "Database operation failed",
		Status:  http.StatusInternalServerError,
	}
}

func NewInternalError(message string) *AppError {
	return &AppError{
		Code:    ErrCodeInternal,
		Message: message,
		Status:  http.StatusInternalServerError,
	}
}

func NewIdempotencyError(message string) *AppError {
	return &AppError{
		Code:    ErrCodeIdempotency,
		Message: message,
		Status:  http.StatusConflict,
	}
}
