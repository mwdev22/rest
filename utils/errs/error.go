package errs

import (
	"fmt"
	"net/http"
)

type ApiError struct {
	StatusCode int    `json:"status_code"`
	Msg        string `json:"msg"`
	ToLog      string `json:"to_log,omitempty"`
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

func InternalServerError(err error) ApiError {
	return ApiError{
		StatusCode: http.StatusInternalServerError,
		Msg:        "internal server error",
		ToLog:      err.Error(),
	}
}

func Unauthorized(err error) ApiError {
	return ApiError{
		StatusCode: http.StatusUnauthorized,
		Msg:        "unauthorized",
		ToLog:      err.Error(),
	}
}

func Forbidden(err error) ApiError {
	return ApiError{
		StatusCode: http.StatusForbidden,
		Msg:        "forbidden",
		ToLog:      err.Error(),
	}
}

func InvalidJson(err error) ApiError {
	return ApiError{
		StatusCode: http.StatusBadRequest,
		Msg:        "invalid json",
		ToLog:      err.Error(),
	}
}

func InvalidFormData(err error) ApiError {
	return ApiError{
		StatusCode: http.StatusBadRequest,
		Msg:        "invalid form data",
		ToLog:      err.Error(),
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

func NotFound(err error) ApiError {
	return ApiError{
		StatusCode: http.StatusNotFound,
		Msg:        "not found",
		ToLog:      err.Error(),
	}
}

func ObjectNotFound(id string, name string) ApiError {
	return ApiError{
		StatusCode: http.StatusNotFound,
		Msg:        fmt.Sprintf("%s with ID %s not found", name, id),
		ToLog:      "",
	}
}
