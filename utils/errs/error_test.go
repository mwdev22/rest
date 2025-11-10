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

func TestInvalidJson(t *testing.T) {
	tests := []struct {
		name        string
		inputError  error
		expectedMsg string
	}{
		{
			name:        "json parse error",
			inputError:  errors.New("unexpected end of JSON input"),
			expectedMsg: "invalid json",
		},
		{
			name:        "nil error",
			inputError:  nil,
			expectedMsg: "invalid json",
		},
		{
			name:        "custom error",
			inputError:  errors.New("some other error"),
			expectedMsg: "invalid json",
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
		name        string
		msg         string
		expectedMsg string
	}{
		{
			name:        "user not found",
			msg:         "user not found",
			expectedMsg: "user not found",
		},
		{
			name:        "resource not found",
			msg:         "resource not found",
			expectedMsg: "resource not found",
		},
		{
			name:        "empty message",
			msg:         "",
			expectedMsg: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NotFound(tt.msg)

			if err.StatusCode != http.StatusNotFound {
				t.Errorf("expected status %d, got %d", http.StatusNotFound, err.StatusCode)
			}
			if err.Msg != tt.expectedMsg {
				t.Errorf("expected msg '%s', got '%s'", tt.expectedMsg, err.Msg)
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
			resource:    "resource",
			expectedMsg: "resource with ID 456 not found",
		},
		{
			name:        "empty message",
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
		})
	}
}
