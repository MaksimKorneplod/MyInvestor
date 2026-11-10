package handler

import (
	"context"
	"errors"
	"github.com/MaksimKorneplod/MyInvestor/internal/service"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

type registrationStore struct{ err error }

func (s registrationStore) CreateSchool(_ context.Context, login, email, hash string) (service.School, error) {
	return service.School{ID: "school-id", Login: login, Email: email}, s.err
}

type blockingRegistrationStore struct {
	started chan<- struct{}
	release <-chan struct{}
}

func (s blockingRegistrationStore) CreateSchool(ctx context.Context, login, email, _ string) (service.School, error) {
	s.started <- struct{}{}
	select {
	case <-s.release:
		return service.School{ID: "school-id", Login: login, Email: email}, nil
	case <-ctx.Done():
		return service.School{}, ctx.Err()
	}
}

func TestSchoolRegistrationRejectsOverload(t *testing.T) {
	started := make(chan struct{}, 3)
	release := make(chan struct{})
	defer close(release)
	handler := SchoolRegistration(service.Registration{Store: blockingRegistrationStore{started, release}})
	request := func() int {
		rec := httptest.NewRecorder()
		body := strings.NewReader(`{"login":"school","email":"school@example.com","password":"password123"}`)
		handler.ServeHTTP(rec, httptest.NewRequest("POST", "/api/v1/schools/register", body))
		return rec.Code
	}
	for range 2 {
		go request()
		select {
		case <-started:
		case <-time.After(5 * time.Second):
			t.Fatal("registration did not reach the store")
		}
	}
	done := make(chan int, 1)
	go func() { done <- request() }()
	select {
	case status := <-done:
		if status != http.StatusServiceUnavailable {
			t.Fatalf("overloaded registration returned %d, want 503", status)
		}
	case <-started:
		t.Fatal("third registration reached the store while both slots were occupied")
	case <-time.After(5 * time.Second):
		t.Fatal("overloaded registration did not return")
	}
}

func TestSchoolRegistrationAPI(t *testing.T) {
	valid := `{"login":"school","email":"school@example.com","password":"password123"}`
	for _, tc := range []struct {
		name, body string
		err        error
		status     int
	}{
		{"created", valid, nil, 201}, {"duplicate", valid, service.ErrAccountExists, 409},
		{"database failure", valid, errors.New("private database details"), 500},
		{"missing fields", `{}`, nil, 400}, {"invalid email", `{"login":"school","email":"bad","password":"password123"}`, nil, 400},
		{"null in login", `{"login":"school\u0000","email":"school@example.com","password":"password123"}`, nil, 400},
		{"malformed", `{`, nil, 400}, {"unknown field", `{"unknown":true}`, nil, 400},
		{"multiple objects", valid + `{}`, nil, 400}, {"oversized", strings.Repeat("x", 9000), nil, 400},
	} {
		t.Run(tc.name, func(t *testing.T) {
			mux := http.NewServeMux()
			mux.HandleFunc("POST /api/v1/schools/register", SchoolRegistration(service.Registration{Store: registrationStore{err: tc.err}}))
			rec := httptest.NewRecorder()
			mux.ServeHTTP(rec, httptest.NewRequest("POST", "/api/v1/schools/register", strings.NewReader(tc.body)))
			if rec.Code != tc.status {
				t.Fatalf("status %d, want %d: %s", rec.Code, tc.status, rec.Body)
			}
			if strings.Contains(rec.Body.String(), "password") && tc.status == 201 {
				t.Fatal("password leaked")
			}
			if strings.Contains(rec.Body.String(), "private database") {
				t.Fatal("internal error leaked")
			}
		})
	}
}
