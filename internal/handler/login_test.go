package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/MaksimKorneplod/MyInvestor/internal/service"
)

type loginStoreStub struct {
	credentials service.SchoolCredentials
	findErr     error
	sessionErr  error
	savedID     string
	savedHash   string
	expiresAt   time.Time
}

type blockingLoginStore struct {
	started chan<- struct{}
	release <-chan struct{}
}

func (s blockingLoginStore) FindSchool(ctx context.Context, _ string) (service.SchoolCredentials, error) {
	s.started <- struct{}{}
	select {
	case <-s.release:
		return service.SchoolCredentials{ID: "school-id", Hash: "invalid hash"}, nil
	case <-ctx.Done():
		return service.SchoolCredentials{}, ctx.Err()
	}
}

func (blockingLoginStore) CreateSchoolSession(context.Context, string, string, time.Time) error {
	return nil
}

func (s *loginStoreStub) FindSchool(context.Context, string) (service.SchoolCredentials, error) {
	return s.credentials, s.findErr
}

func (s *loginStoreStub) CreateSchoolSession(_ context.Context, schoolID, tokenHash string, expiresAt time.Time) error {
	s.savedID = schoolID
	s.savedHash = tokenHash
	s.expiresAt = expiresAt
	return s.sessionErr
}

func TestSchoolLoginSuccess(t *testing.T) {
	password := "correct horse battery staple"
	hash, err := service.HashPassword(password)
	if err != nil {
		t.Fatal(err)
	}
	store := &loginStoreStub{credentials: service.SchoolCredentials{ID: "school-id", Hash: hash}}
	handler := SchoolLogin(service.Login{Store: store})
	request := httptest.NewRequest(http.MethodPost, "/api/v1/schools/login",
		strings.NewReader(`{"email":"school@example.com","password":"correct horse battery staple"}`))
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d: %s", recorder.Code, http.StatusOK, recorder.Body)
	}
	var response struct {
		AccessToken string `json:"accessToken"`
		TokenType   string `json:"tokenType"`
		ExpiresIn   int64  `json:"expiresIn"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatal(err)
	}
	if response.AccessToken == "" || response.TokenType != "Bearer" || response.ExpiresIn != 86400 {
		t.Fatalf("unexpected login response: %+v", response)
	}
	if store.savedID != "school-id" || store.savedHash == "" || store.expiresAt.IsZero() {
		t.Fatalf("session was not persisted: %+v", store)
	}
	if strings.Contains(recorder.Body.String(), store.savedHash) {
		t.Fatal("response contains stored token hash")
	}
}

func TestSchoolLoginInvalidCredentialsAreIndistinguishable(t *testing.T) {
	password := "correct horse battery staple"
	hash, err := service.HashPassword(password)
	if err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		name  string
		store *loginStoreStub
		body  string
	}{
		{
			name:  "unknown email",
			store: &loginStoreStub{findErr: service.ErrAccountNotFound},
			body:  `{"email":"unknown@example.com","password":"password"}`,
		},
		{
			name:  "wrong password",
			store: &loginStoreStub{credentials: service.SchoolCredentials{ID: "school-id", Hash: hash}},
			body:  `{"email":"school@example.com","password":"wrong password"}`,
		},
	}
	var firstBody string
	for i, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			SchoolLogin(service.Login{Store: tt.store}).ServeHTTP(
				recorder,
				httptest.NewRequest(http.MethodPost, "/api/v1/schools/login", strings.NewReader(tt.body)),
			)
			if recorder.Code != http.StatusUnauthorized {
				t.Fatalf("status = %d, want %d", recorder.Code, http.StatusUnauthorized)
			}
			if i == 0 {
				firstBody = recorder.Body.String()
			} else if recorder.Body.String() != firstBody {
				t.Fatalf("invalid-credential response differs: got %q, want %q", recorder.Body.String(), firstBody)
			}
			if tt.store.savedID != "" {
				t.Fatal("session was created for invalid credentials")
			}
		})
	}
}

func TestSchoolLoginRejectsInvalidRequest(t *testing.T) {
	valid := `{"email":"school@example.com","password":"password"}`
	for _, body := range []string{
		`{`,
		`{}`,
		`{"email":"school@example.com","password":"password","unexpected":true}`,
		valid + valid,
		strings.Repeat("x", 9000),
	} {
		recorder := httptest.NewRecorder()
		SchoolLogin(service.Login{Store: &loginStoreStub{}}).ServeHTTP(
			recorder,
			httptest.NewRequest(http.MethodPost, "/api/v1/schools/login", strings.NewReader(body)),
		)
		if recorder.Code != http.StatusBadRequest {
			t.Errorf("body %q: status = %d, want %d", body[:min(len(body), 40)], recorder.Code, http.StatusBadRequest)
		}
	}
}

func TestSchoolLoginDoesNotExposeStoreErrors(t *testing.T) {
	recorder := httptest.NewRecorder()
	SchoolLogin(service.Login{Store: &loginStoreStub{findErr: errors.New("private database details")}}).ServeHTTP(
		recorder,
		httptest.NewRequest(http.MethodPost, "/api/v1/schools/login",
			strings.NewReader(`{"email":"school@example.com","password":"password"}`)),
	)
	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusInternalServerError)
	}
	if strings.Contains(recorder.Body.String(), "private database details") {
		t.Fatal("internal error leaked in response")
	}
}

func TestSchoolLoginRejectsOverload(t *testing.T) {
	started := make(chan struct{}, 2)
	release := make(chan struct{})
	var releaseOnce sync.Once
	defer func() { releaseOnce.Do(func() { close(release) }) }()
	loginHandler := SchoolLogin(service.Login{Store: blockingLoginStore{started: started, release: release}})
	body := `{"email":"school@example.com","password":"password"}`
	done := make(chan int, 2)

	for range 2 {
		go func() {
			recorder := httptest.NewRecorder()
			loginHandler.ServeHTTP(recorder, httptest.NewRequest(http.MethodPost, "/api/v1/schools/login", strings.NewReader(body)))
			done <- recorder.Code
		}()
		select {
		case <-started:
		case <-time.After(5 * time.Second):
			t.Fatal("login did not reach the store")
		}
	}

	recorder := httptest.NewRecorder()
	loginHandler.ServeHTTP(recorder, httptest.NewRequest(http.MethodPost, "/api/v1/schools/login", strings.NewReader(body)))
	if recorder.Code != http.StatusServiceUnavailable {
		t.Fatalf("overloaded login returned %d, want %d", recorder.Code, http.StatusServiceUnavailable)
	}
	if !strings.Contains(recorder.Body.String(), "Too many login attempts") {
		t.Fatalf("unexpected overload response: %s", recorder.Body)
	}

	releaseOnce.Do(func() { close(release) })
	for range 2 {
		select {
		case <-done:
		case <-time.After(5 * time.Second):
			t.Fatal("in-flight login did not finish")
		}
	}
}
