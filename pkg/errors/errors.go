package errors

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// ErrorCode represents a standard error code for API responses
type ErrorCode string

const (
	// Client errors (4xx)
	ErrCodeBadRequest          ErrorCode = "BAD_REQUEST"
	ErrCodeUnauthorized        ErrorCode = "UNAUTHORIZED"
	ErrCodeForbidden           ErrorCode = "FORBIDDEN"
	ErrCodeNotFound            ErrorCode = "NOT_FOUND"
	ErrCodeConflict            ErrorCode = "CONFLICT"
	ErrCodeValidation          ErrorCode = "VALIDATION_ERROR"
	ErrCodeRateLimitExceeded   ErrorCode = "RATE_LIMIT_EXCEEDED"
	ErrCodeInvalidJSON         ErrorCode = "INVALID_JSON"
	ErrCodeMissingField        ErrorCode = "MISSING_FIELD"
	
	// Server errors (5xx)
	ErrCodeInternalServer      ErrorCode = "INTERNAL_SERVER_ERROR"
	ErrCodeDatabaseError       ErrorCode = "DATABASE_ERROR"
	ErrCodeExternalService     ErrorCode = "EXTERNAL_SERVICE_ERROR"
)

// APIError represents a structured error response for the API
type APIError struct {
	Code       ErrorCode `json:"code"`
	Message    string    `json:"message"`
	Details    string    `json:"details,omitempty"`
	RequestID  string    `json:"request_id,omitempty"`
	StatusCode int       `json:"-"`
}

// Error implements the error interface
func (e *APIError) Error() string {
	if e.Details != "" {
		return fmt.Sprintf("%s: %s - %s", e.Code, e.Message, e.Details)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// NewAPIError creates a new APIError with the given parameters
func NewAPIError(code ErrorCode, message string, statusCode int) *APIError {
	return &APIError{
		Code:       code,
		Message:    message,
		StatusCode: statusCode,
	}
}

// WithDetails adds details to an existing APIError
func (e *APIError) WithDetails(details string) *APIError {
	e.Details = details
	return e
}

// WithRequestID adds a request ID to an existing APIError
func (e *APIError) WithRequestID(requestID string) *APIError {
	e.RequestID = requestID
	return e
}

// Common error constructors
func BadRequest(message string) *APIError {
	return NewAPIError(ErrCodeBadRequest, message, http.StatusBadRequest)
}

func Unauthorized(message string) *APIError {
	return NewAPIError(ErrCodeUnauthorized, message, http.StatusUnauthorized)
}

func Forbidden(message string) *APIError {
	return NewAPIError(ErrCodeForbidden, message, http.StatusForbidden)
}

func NotFound(message string) *APIError {
	return NewAPIError(ErrCodeNotFound, message, http.StatusNotFound)
}

func Conflict(message string) *APIError {
	return NewAPIError(ErrCodeConflict, message, http.StatusConflict)
}

func ValidationError(message string) *APIError {
	return NewAPIError(ErrCodeValidation, message, http.StatusBadRequest)
}

func RateLimitExceeded(message string) *APIError {
	return NewAPIError(ErrCodeRateLimitExceeded, message, http.StatusTooManyRequests)
}

func InvalidJSON(message string) *APIError {
	return NewAPIError(ErrCodeInvalidJSON, message, http.StatusBadRequest)
}

func InternalServerError(message string) *APIError {
	return NewAPIError(ErrCodeInternalServer, message, http.StatusInternalServerError)
}

func DatabaseError(message string) *APIError {
	return NewAPIError(ErrCodeDatabaseError, message, http.StatusInternalServerError)
}

func ExternalServiceError(message string) *APIError {
	return NewAPIError(ErrCodeExternalService, message, http.StatusInternalServerError)
}

// WriteError writes an APIError as a JSON response
func WriteError(w http.ResponseWriter, err *APIError) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(err.StatusCode)
	json.NewEncoder(w).Encode(err)
}

// WriteErrorWithRequestID writes an APIError with a request ID as a JSON response
func WriteErrorWithRequestID(w http.ResponseWriter, err *APIError, requestID string) {
	err.RequestID = requestID
	WriteError(w, err)
}
