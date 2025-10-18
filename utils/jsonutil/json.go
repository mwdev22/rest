package jsonutil

import (
	"encoding/json"
	"net/http"

	"github.com/go-playground/validator"
)

var Validate = validator.New()

func Write(w http.ResponseWriter, status int, body any) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(body); err != nil {
		return err
	}
	return nil
}

func Parse(r *http.Request, payload any) error {
	defer r.Body.Close()
	if err := json.NewDecoder(r.Body).Decode(payload); err != nil {
		return err
	}
	if err := Validate.Struct(payload); err != nil {
		return err
	}
	return nil
}
