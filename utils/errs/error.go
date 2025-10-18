package errs

import "fmt"

type ApiError struct {
	StatusCode int `json:"status_code"`
	Msg        any `json:"msg"`
}

func (e ApiError) Error() string {
	return fmt.Sprintf("ERROR: %v", e.Msg)
}

func NewApiError(status int, msg string) *ApiError {
	return &ApiError{
		StatusCode: status,
		Msg:        msg,
	}
}
