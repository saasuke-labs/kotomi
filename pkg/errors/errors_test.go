package errors

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAPIError_Error(t *testing.T) {
	tests := []struct {
		name     string
		err      *APIError
		expected string
	}{
		{
			name:     "Error without details",
			err:      &APIError{Code: ErrCodeBadRequest, Message: "Invalid input"},
			expected: "BAD_REQUEST: Invalid input",
		},
		{
			name:     "Error with details",
			err:      &APIError{Code: ErrCodeBadRequest, Message: "Invalid input", Details: "Field 'name' is required"},
			expected: "BAD_REQUEST: Invalid input - Field 'name' is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.Error(); got != tt.expected {
				t.Errorf("Error() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestNewAPIError(t *testing.T) {
	err := NewAPIError(ErrCodeBadRequest, "Test message", http.StatusBadRequest)
	
	if err.Code != ErrCodeBadRequest {
		t.Errorf("Expected code %v, got %v", ErrCodeBadRequest, err.Code)
	}
	if err.Message != "Test message" {
		t.Errorf("Expected message 'Test message', got %v", err.Message)
	}
	if err.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status %v, got %v", http.StatusBadRequest, err.StatusCode)
	}
}

func TestAPIError_WithDetails(t *testing.T) {
	err := BadRequest("Test message").WithDetails("Additional details")
	
	if err.Details != "Additional details" {
		t.Errorf("Expected details 'Additional details', got %v", err.Details)
	}
}

func TestAPIError_WithRequestID(t *testing.T) {
	requestID := "test-request-id-123"
	err := BadRequest("Test message").WithRequestID(requestID)
	
	if err.RequestID != requestID {
		t.Errorf("Expected request ID %v, got %v", requestID, err.RequestID)
	}
}

func TestCommonErrorConstructors(t *testing.T) {
	tests := []struct {
		name       string
		constructor func(string) *APIError
		code       ErrorCode
		status     int
	}{
		{"BadRequest", BadRequest, ErrCodeBadRequest, http.StatusBadRequest},
		{"Unauthorized", Unauthorized, ErrCodeUnauthorized, http.StatusUnauthorized},
		{"Forbidden", Forbidden, ErrCodeForbidden, http.StatusForbidden},
		{"NotFound", NotFound, ErrCodeNotFound, http.StatusNotFound},
		{"Conflict", Conflict, ErrCodeConflict, http.StatusConflict},
		{"ValidationError", ValidationError, ErrCodeValidation, http.StatusBadRequest},
		{"RateLimitExceeded", RateLimitExceeded, ErrCodeRateLimitExceeded, http.StatusTooManyRequests},
		{"InvalidJSON", InvalidJSON, ErrCodeInvalidJSON, http.StatusBadRequest},
		{"InternalServerError", InternalServerError, ErrCodeInternalServer, http.StatusInternalServerError},
		{"DatabaseError", DatabaseError, ErrCodeDatabaseError, http.StatusInternalServerError},
		{"ExternalServiceError", ExternalServiceError, ErrCodeExternalService, http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.constructor("Test message")
			
			if err.Code != tt.code {
				t.Errorf("Expected code %v, got %v", tt.code, err.Code)
			}
			if err.Message != "Test message" {
				t.Errorf("Expected message 'Test message', got %v", err.Message)
			}
			if err.StatusCode != tt.status {
				t.Errorf("Expected status %v, got %v", tt.status, err.StatusCode)
			}
		})
	}
}

func TestWriteError(t *testing.T) {
	err := BadRequest("Test error message")
	w := httptest.NewRecorder()
	
	WriteError(w, err)
	
	// Check status code
	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status code %v, got %v", http.StatusBadRequest, w.Code)
	}
	
	// Check content type
	contentType := w.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("Expected Content-Type 'application/json', got %v", contentType)
	}
	
	// Check JSON response
	var response APIError
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}
	
	if response.Code != ErrCodeBadRequest {
		t.Errorf("Expected code %v, got %v", ErrCodeBadRequest, response.Code)
	}
	if response.Message != "Test error message" {
		t.Errorf("Expected message 'Test error message', got %v", response.Message)
	}
}

func TestWriteErrorWithRequestID(t *testing.T) {
	err := BadRequest("Test error message")
	requestID := "test-request-123"
	w := httptest.NewRecorder()
	
	WriteErrorWithRequestID(w, err, requestID)
	
	// Check JSON response includes request ID
	var response APIError
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}
	
	if response.RequestID != requestID {
		t.Errorf("Expected request ID %v, got %v", requestID, response.RequestID)
	}
	if response.Code != ErrCodeBadRequest {
		t.Errorf("Expected code %v, got %v", ErrCodeBadRequest, response.Code)
	}
}

func TestWriteError_WithDetails(t *testing.T) {
	err := BadRequest("Invalid input").WithDetails("Field 'name' is required")
	w := httptest.NewRecorder()
	
	WriteError(w, err)
	
	var response APIError
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}
	
	if response.Details != "Field 'name' is required" {
		t.Errorf("Expected details 'Field 'name' is required', got %v", response.Details)
	}
}
