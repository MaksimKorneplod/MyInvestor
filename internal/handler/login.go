package handler

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"strings"

	"github.com/MaksimKorneplod/MyInvestor/internal/service"
)

func SchoolLogin(s service.Login) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		fail := func(status int, message string) {
			w.WriteHeader(status)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": message})
		}

		var input struct {
			Email    string `json:"email"`
			Password string `json:"password"`
		}
		r.Body = http.MaxBytesReader(w, r.Body, 8192)
		decoder := json.NewDecoder(r.Body)
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&input); err != nil {
			fail(http.StatusBadRequest, "Invalid request.")
			return
		}
		if err := decoder.Decode(new(any)); err != io.EOF {
			fail(http.StatusBadRequest, "Invalid request.")
			return
		}
		if strings.TrimSpace(input.Email) == "" || input.Password == "" {
			fail(http.StatusBadRequest, "Email and password are required.")
			return
		}

		session, err := s.Authenticate(r.Context(), input.Email, input.Password)
		switch {
		case errors.Is(err, service.ErrInvalidCredentials):
			fail(http.StatusUnauthorized, "Invalid email or password.")
		case err != nil:
			if r.Context().Err() != nil {
				return
			}
			slog.Error("school login failed", "error", err)
			fail(http.StatusInternalServerError, "Internal server error.")
		default:
			w.WriteHeader(http.StatusOK)
			if err := json.NewEncoder(w).Encode(session); err != nil {
				slog.Error("failed to encode school login response", "error", err)
			}
		}
	}
}
