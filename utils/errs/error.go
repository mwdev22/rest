package errs

import (
	"fmt"
	"net/http"
)

type ApiError struct {
	StatusCode int    `json:"status_code"`
	Msg        string `json:"msg"`
}

func (e ApiError) Error() string {
	return e.Msg
}

func (e ApiError) Map() map[string]string {
	return map[string]string{
		"error": e.Error(),
	}
}

func NewApiError(status int, msg string) ApiError {
	return ApiError{
		StatusCode: status,
		Msg:        msg,
	}
}

func InternalServerError() ApiError {
	return ApiError{
		StatusCode: http.StatusInternalServerError,
		Msg:        "internal server error",
	}
}

func InvalidJson() ApiError {
	return ApiError{
		StatusCode: http.StatusBadRequest,
		Msg:        "invalid json",
	}
}

func InvalidFormData() ApiError {
	return ApiError{
		StatusCode: http.StatusBadRequest,
		Msg:        "invalid form data",
	}
}

func InvalidPathParam(param string) ApiError {
	return ApiError{
		StatusCode: http.StatusBadRequest,
		Msg:        fmt.Sprintf("invalid path param: %s", param),
	}
}

func InvalidQueryParam(param string) ApiError {
	return ApiError{
		StatusCode: http.StatusBadRequest,
		Msg:        fmt.Sprintf("invalid query param: %s", param),
	}
}

func NotFound(msg string) ApiError {
	return ApiError{
		StatusCode: http.StatusNotFound,
		Msg:        msg,
	}
}

func ObjectNotFound(id string, name string) ApiError {
	return ApiError{
		StatusCode: http.StatusNotFound,
		Msg:        fmt.Sprintf("%s with ID %s not found", name, id),
	}
}
