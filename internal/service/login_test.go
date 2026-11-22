package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"testing"
	"time"
)

type emailStoreStub struct {
	email       string
	credentials SchoolCredentials
	err         error
	sessionID   string
	tokenHash   string
	expiresAt   time.Time
	sessionErr  error
}

func (s *emailStoreStub) FindSchool(_ context.Context, email string) (SchoolCredentials, error) {
	s.email = email
	return s.credentials, s.err
}

func (s *emailStoreStub) CreateSchoolSession(_ context.Context, schoolID, tokenHash string, expiresAt time.Time) error {
	s.sessionID = schoolID
	s.tokenHash = tokenHash
	s.expiresAt = expiresAt
	return s.sessionErr
}

func TestLoginAuthenticate(t *testing.T) {
	password := "correct horse battery staple"
	hash, err := HashPassword(password)
	if err != nil {
		t.Fatal(err)
	}

	store := &emailStoreStub{
		credentials: SchoolCredentials{ID: "school-id", Hash: hash},
	}
	got, err := (Login{Store: store}).Authenticate(context.Background(), " SCHOOL-EMAIL ", password)
	if err != nil {
		t.Fatalf("Authenticate() error = %v", err)
	}
	if got.TokenType != "Bearer" || got.ExpiresIn != 86400 || got.AccessToken == "" {
		t.Fatalf("Authenticate() session = %+v", got)
	}
	if store.email != "school-email" {
		t.Fatalf("FindSchool() email = %q, want normalized email", store.email)
	}
	if store.sessionID != "school-id" || store.tokenHash == "" || store.expiresAt.IsZero() {
		t.Fatalf("CreateSchoolSession() values: id=%q hash=%q expires=%s", store.sessionID, store.tokenHash, store.expiresAt)
	}
	sum := sha256.Sum256([]byte(got.AccessToken))
	if store.tokenHash != hex.EncodeToString(sum[:]) {
		t.Fatal("stored token hash does not match access token")
	}
}

func TestLoginAuthenticateInvalidCredentials(t *testing.T) {
	password := "correct horse battery staple"
	hash, err := HashPassword(password)
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name     string
		store    *emailStoreStub
		password string
	}{
		{
			name: "wrong password",
			store: &emailStoreStub{
				credentials: SchoolCredentials{ID: "school-id", Hash: hash},
			},
			password: "wrong password",
		},
		{
			name:     "school not found",
			store:    &emailStoreStub{err: ErrAccountNotFound},
			password: password,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := (Login{Store: tt.store}).Authenticate(context.Background(), "school-email", tt.password)
			if !errors.Is(err, ErrInvalidCredentials) {
				t.Fatalf("Authenticate() error = %v, want %v", err, ErrInvalidCredentials)
			}
			if tt.store.sessionID != "" {
				t.Fatal("session was created for invalid credentials")
			}
		})
	}
}

func TestDummyPasswordHashIsWellFormed(t *testing.T) {
	if err := VerifyPassword("not-the-dummy-password", dummyPasswordHash); !errors.Is(err, ErrPasswordMismatch) {
		t.Fatalf("VerifyPassword() error = %v, want password mismatch for valid dummy hash", err)
	}
}

func TestLoginAuthenticateStoreError(t *testing.T) {
	storeErr := errors.New("database unavailable")
	store := &emailStoreStub{err: storeErr}

	_, err := (Login{Store: store}).Authenticate(context.Background(), "school-email", "password")
	if !errors.Is(err, storeErr) {
		t.Fatalf("Authenticate() error = %v, want wrapped %v", err, storeErr)
	}
}

func TestLoginAuthenticateInvalidStoredHash(t *testing.T) {
	store := &emailStoreStub{
		credentials: SchoolCredentials{ID: "school-id", Hash: "invalid hash"},
	}

	_, err := (Login{Store: store}).Authenticate(context.Background(), "school-email", "password")
	if !errors.Is(err, ErrInvalidPasswordHash) {
		t.Fatalf("Authenticate() error = %v, want %v", err, ErrInvalidPasswordHash)
	}
}

func TestLoginAuthenticateSessionStoreError(t *testing.T) {
	password := "correct horse battery staple"
	hash, err := HashPassword(password)
	if err != nil {
		t.Fatal(err)
	}
	storeErr := errors.New("session database unavailable")
	store := &emailStoreStub{
		credentials: SchoolCredentials{ID: "school-id", Hash: hash},
		sessionErr:  storeErr,
	}

	if _, err := (Login{Store: store}).Authenticate(context.Background(), "school-email", password); !errors.Is(err, storeErr) {
		t.Fatalf("Authenticate() error = %v, want wrapped %v", err, storeErr)
	}
}
