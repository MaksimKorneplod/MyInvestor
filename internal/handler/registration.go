package handler

import (
	"encoding/json"
	"errors"
	"github.com/MaksimKorneplod/MyInvestor/internal/service"
	"io"
	"log/slog"
	"net/http"
)

func SchoolRegistration(s service.Registration) http.HandlerFunc {
	// Bound concurrent work: password hashing is deliberately CPU-intensive.
	inFlight := make(chan struct{}, 2)
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fail := func(status int, code, message string) {
			w.WriteHeader(status)
			_ = json.NewEncoder(w).Encode(map[string]any{"error": map[string]string{"code": code, "message": message}})
		}
		var input struct {
			Login    string `json:"login"`
			Email    string `json:"email"`
			Password string `json:"password"`
		}
		r.Body = http.MaxBytesReader(w, r.Body, 8192)
		decoder := json.NewDecoder(r.Body)
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&input); err != nil {
			fail(400, "invalid_request", "Expected a JSON object with login, email and password")
			return
		}
		if err := decoder.Decode(new(any)); err != io.EOF {
			fail(400, "invalid_request", "Expected exactly one JSON object")
			return
		}
		select {
		case inFlight <- struct{}{}:
			defer func() { <-inFlight }()
		default:
			fail(http.StatusServiceUnavailable, "registration_busy", "Too many registrations; retry later")
			return
		}
		school, err := s.Register(r.Context(), input.Login, input.Email, input.Password)
		switch {
		case errors.Is(err, service.ErrInvalidRegistration):
			fail(400, "validation_error", err.Error())
		case errors.Is(err, service.ErrAccountExists):
			fail(409, "account_exists", "Email or login already registered")
		case err != nil:
			slog.Error("school registration failed", "error", err)
			fail(500, "internal_error", "Unable to register school")
		default:
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(school)
		}
	}
}
