package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/oklog/ulid/v2"
)

func encode[T any](w http.ResponseWriter, r *http.Request, status int, v T) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		return fmt.Errorf("encode json: %w", err)
	}
	return nil
}

func renderErr(w http.ResponseWriter, r *http.Request, err error) error {
	return encode(w, r, status(err), errRes{Error: err.Error()})
}

func status(err error) int {
	if err == nil {
		return http.StatusOK
	}

	if errors.Is(err, ulid.ErrDataSize) {
		return http.StatusBadRequest
	}

	if errors.Is(err, ulid.ErrInvalidCharacters) {
		return http.StatusBadRequest
	}

	return http.StatusInternalServerError
}
