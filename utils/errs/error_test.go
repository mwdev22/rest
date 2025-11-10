package errs

import (
	"errors"
	"net/http"
	"testing"
)

func TestNewApiError(t *testing.T) {
	tests := []struct {
		name           string
		status         int
		msg            string
		expectedStatus int
		expectedMsg    string
	}{
		{
			name:           "bad request error",
			status:         http.StatusBadRequest,
			msg:            "invalid input",
			expectedStatus: http.StatusBadRequest,
			expectedMsg:    "invalid input",
		},
		{
			name:           "not found error",
			status:         http.StatusNotFound,
			msg:            "resource not found",
			expectedStatus: http.StatusNotFound,
			expectedMsg:    "resource not found",
		},
		{
			name:           "internal server error",
			status:         http.StatusInternalServerError,
			msg:            "something went wrong",
			expectedStatus: http.StatusInternalServerError,
			expectedMsg:    "something went wrong",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NewApiError(tt.status, tt.msg)

			if err.StatusCode != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, err.StatusCode)
			}
			if err.Msg != tt.expectedMsg {
				t.Errorf("expected msg '%s', got '%s'", tt.expectedMsg, err.Msg)
			}
			if err.ToLog != "" {
				t.Errorf("expected empty ToLog, got '%s'", err.ToLog)
			}
		})
	}
}

func TestApiError_Error(t *testing.T) {
	tests := []struct {
		name     string
		apiError ApiError
		expected string
	}{
		{
			name:     "simple error message",
			apiError: NewApiError(http.StatusBadRequest, "bad request"),
			expected: "bad request",
		},
		{
			name:     "error with special characters",
			apiError: NewApiError(http.StatusNotFound, "user: not found!"),
			expected: "user: not found!",
		},
		{
			name:     "empty message",
			apiError: NewApiError(http.StatusInternalServerError, ""),
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.apiError.Error()
			if result != tt.expected {
				t.Errorf("expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestApiError_Map(t *testing.T) {
	tests := []struct {
		name          string
		apiError      ApiError
		expectedError string
	}{
		{
			name:          "simple error",
			apiError:      NewApiError(http.StatusBadRequest, "validation failed"),
			expectedError: "validation failed",
		},
		{
			name:          "another error",
			apiError:      NewApiError(http.StatusUnauthorized, "unauthorized access"),
			expectedError: "unauthorized access",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := tt.apiError.Map()

			if len(m) != 1 {
				t.Errorf("expected map with 1 key, got %d", len(m))
			}

			if m["error"] != tt.expectedError {
				t.Errorf("expected error '%s', got '%s'", tt.expectedError, m["error"])
			}
		})
	}
}

func TestInternalServerError(t *testing.T) {
	tests := []struct {
		name          string
		inputError    error
		expectedMsg   string
		expectedToLog string
	}{
		{
			name:          "database error",
			inputError:    errors.New("connection timeout"),
			expectedMsg:   "internal server error",
			expectedToLog: "connection timeout",
		},
		{
			name:          "generic error",
			inputError:    errors.New("unexpected error"),
			expectedMsg:   "internal server error",
			expectedToLog: "unexpected error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := InternalServerError(tt.inputError)

			if err.StatusCode != http.StatusInternalServerError {
				t.Errorf("expected status %d, got %d", http.StatusInternalServerError, err.StatusCode)
			}
			if err.Msg != tt.expectedMsg {
				t.Errorf("expected msg '%s', got '%s'", tt.expectedMsg, err.Msg)
			}
			if err.ToLog != tt.expectedToLog {
				t.Errorf("expected ToLog '%s', got '%s'", tt.expectedToLog, err.ToLog)
			}
		})
	}
}

func TestUnauthorized(t *testing.T) {
	tests := []struct {
		name          string
		inputError    error
		expectedMsg   string
		expectedToLog string
	}{
		{
			name:          "invalid token",
			inputError:    errors.New("token expired"),
			expectedMsg:   "unauthorized",
			expectedToLog: "token expired",
		},
		{
			name:          "missing credentials",
			inputError:    errors.New("no credentials provided"),
			expectedMsg:   "unauthorized",
			expectedToLog: "no credentials provided",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Unauthorized(tt.inputError)

			if err.StatusCode != http.StatusUnauthorized {
				t.Errorf("expected status %d, got %d", http.StatusUnauthorized, err.StatusCode)
			}
			if err.Msg != tt.expectedMsg {
				t.Errorf("expected msg '%s', got '%s'", tt.expectedMsg, err.Msg)
			}
			if err.ToLog != tt.expectedToLog {
				t.Errorf("expected ToLog '%s', got '%s'", tt.expectedToLog, err.ToLog)
			}
		})
	}
}

func TestForbidden(t *testing.T) {
	tests := []struct {
		name          string
		inputError    error
		expectedMsg   string
		expectedToLog string
	}{
		{
			name:          "insufficient permissions",
			inputError:    errors.New("user lacks required role"),
			expectedMsg:   "forbidden",
			expectedToLog: "user lacks required role",
		},
		{
			name:          "access denied",
			inputError:    errors.New("resource access denied"),
			expectedMsg:   "forbidden",
			expectedToLog: "resource access denied",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Forbidden(tt.inputError)

			if err.StatusCode != http.StatusForbidden {
				t.Errorf("expected status %d, got %d", http.StatusForbidden, err.StatusCode)
			}
			if err.Msg != tt.expectedMsg {
				t.Errorf("expected msg '%s', got '%s'", tt.expectedMsg, err.Msg)
			}
			if err.ToLog != tt.expectedToLog {
				t.Errorf("expected ToLog '%s', got '%s'", tt.expectedToLog, err.ToLog)
			}
		})
	}
}

func TestInvalidJson(t *testing.T) {
	tests := []struct {
		name          string
		inputError    error
		expectedMsg   string
		expectedToLog string
	}{
		{
			name:          "json parse error",
			inputError:    errors.New("unexpected end of JSON input"),
			expectedMsg:   "invalid json",
			expectedToLog: "unexpected end of JSON input",
		},
		{
			name:          "syntax error",
			inputError:    errors.New("invalid character"),
			expectedMsg:   "invalid json",
			expectedToLog: "invalid character",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := InvalidJson(tt.inputError)

			if err.StatusCode != http.StatusBadRequest {
				t.Errorf("expected status %d, got %d", http.StatusBadRequest, err.StatusCode)
			}
			if err.Msg != tt.expectedMsg {
				t.Errorf("expected msg '%s', got '%s'", tt.expectedMsg, err.Msg)
			}
			if err.ToLog != tt.expectedToLog {
				t.Errorf("expected ToLog '%s', got '%s'", tt.expectedToLog, err.ToLog)
			}
		})
	}
}

func TestInvalidFormData(t *testing.T) {
	tests := []struct {
		name          string
		inputError    error
		expectedMsg   string
		expectedToLog string
	}{
		{
			name:          "missing field",
			inputError:    errors.New("required field missing"),
			expectedMsg:   "invalid form data",
			expectedToLog: "required field missing",
		},
		{
			name:          "validation error",
			inputError:    errors.New("email format invalid"),
			expectedMsg:   "invalid form data",
			expectedToLog: "email format invalid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := InvalidFormData(tt.inputError)
			if err.StatusCode != http.StatusBadRequest {
				t.Errorf("expected status %d, got %d", http.StatusBadRequest, err.StatusCode)
			}
			if err.Msg != tt.expectedMsg {
				t.Errorf("expected msg '%s', got '%s'", tt.expectedMsg, err.Msg)
			}
			if err.ToLog != tt.expectedToLog {
				t.Errorf("expected ToLog '%s', got '%s'", tt.expectedToLog, err.ToLog)
			}
		})
	}
}

func TestInvalidPathParam(t *testing.T) {
	tests := []struct {
		name        string
		param       string
		expectedMsg string
	}{
		{
			name:        "id param",
			param:       "id",
			expectedMsg: "invalid path param: id",
		},
		{
			name:        "userId param",
			param:       "userId",
			expectedMsg: "invalid path param: userId",
		},
		{
			name:        "empty param",
			param:       "",
			expectedMsg: "invalid path param: ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := InvalidPathParam(tt.param)

			if err.StatusCode != http.StatusBadRequest {
				t.Errorf("expected status %d, got %d", http.StatusBadRequest, err.StatusCode)
			}
			if err.Msg != tt.expectedMsg {
				t.Errorf("expected msg '%s', got '%s'", tt.expectedMsg, err.Msg)
			}
		})
	}
}

func TestInvalidQueryParam(t *testing.T) {
	tests := []struct {
		name        string
		param       string
		expectedMsg string
	}{
		{
			name:        "limit param",
			param:       "limit",
			expectedMsg: "invalid query param: limit",
		},
		{
			name:        "offset param",
			param:       "offset",
			expectedMsg: "invalid query param: offset",
		},
		{
			name:        "empty param",
			param:       "",
			expectedMsg: "invalid query param: ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := InvalidQueryParam(tt.param)

			if err.StatusCode != http.StatusBadRequest {
				t.Errorf("expected status %d, got %d", http.StatusBadRequest, err.StatusCode)
			}
			if err.Msg != tt.expectedMsg {
				t.Errorf("expected msg '%s', got '%s'", tt.expectedMsg, err.Msg)
			}
		})
	}
}

func TestNotFound(t *testing.T) {
	tests := []struct {
		name          string
		inputError    error
		expectedMsg   string
		expectedToLog string
	}{
		{
			name:          "user not found",
			inputError:    errors.New("user not found in database"),
			expectedMsg:   "not found",
			expectedToLog: "user not found in database",
		},
		{
			name:          "resource not found",
			inputError:    errors.New("resource does not exist"),
			expectedMsg:   "not found",
			expectedToLog: "resource does not exist",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NotFound(tt.inputError)

			if err.StatusCode != http.StatusNotFound {
				t.Errorf("expected status %d, got %d", http.StatusNotFound, err.StatusCode)
			}
			if err.Msg != tt.expectedMsg {
				t.Errorf("expected msg '%s', got '%s'", tt.expectedMsg, err.Msg)
			}
			if err.ToLog != tt.expectedToLog {
				t.Errorf("expected ToLog '%s', got '%s'", tt.expectedToLog, err.ToLog)
			}
		})
	}
}

func TestObjectNotFound(t *testing.T) {
	tests := []struct {
		name        string
		id          string
		resource    string
		expectedMsg string
	}{
		{
			name:        "user not found",
			id:          "123",
			resource:    "user",
			expectedMsg: "user with ID 123 not found",
		},
		{
			name:        "resource not found",
			id:          "456",
			resource:    "product",
			expectedMsg: "product with ID 456 not found",
		},
		{
			name:        "empty id and resource",
			id:          "",
			resource:    "",
			expectedMsg: " with ID  not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ObjectNotFound(tt.id, tt.resource)

			if err.StatusCode != http.StatusNotFound {
				t.Errorf("expected status %d, got %d", http.StatusNotFound, err.StatusCode)
			}
			if err.Msg != tt.expectedMsg {
				t.Errorf("expected msg '%s', got '%s'", tt.expectedMsg, err.Msg)
			}
			if err.ToLog != "" {
				t.Errorf("expected empty ToLog, got '%s'", err.ToLog)
			}
		})
	}
}
