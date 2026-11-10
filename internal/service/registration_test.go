package service

import (
	"context"
	"crypto/pbkdf2"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"strings"
	"testing"
)

type storeStub struct {
	called bool
	hash   string
	err    error
}

func (s *storeStub) CreateSchool(_ context.Context, login, email, hash string) (School, error) {
	s.called = true
	s.hash = hash
	return School{ID: "school-id", Login: login, Email: email}, s.err
}

func TestRegistration(t *testing.T) {
	store := &storeStub{}
	school, err := (Registration{Store: store}).Register(context.Background(), " School ", " SCHOOL@example.com ", "password123")
	if err != nil || school.Email != "school@example.com" || school.Login != "School" {
		t.Fatalf("registration: %+v, %v", school, err)
	}
	parts := strings.Split(store.hash, "$")
	if len(parts) != 4 || parts[0] != "pbkdf2-sha256" || parts[1] != "600000" {
		t.Fatalf("invalid hash format")
	}
	salt, _ := base64.RawStdEncoding.DecodeString(parts[2])
	key, _ := pbkdf2.Key(sha256.New, "password123", salt, 600000, 32)
	if base64.RawStdEncoding.EncodeToString(key) != parts[3] {
		t.Fatal("hash cannot verify password")
	}
	other, err := HashPassword("password123")
	if err != nil || other == store.hash {
		t.Fatal("password hashes must use independent salts")
	}
	store.err = ErrAccountExists
	if _, err := (Registration{Store: store}).Register(context.Background(), "School", "school@example.com", "password123"); !errors.Is(err, ErrAccountExists) {
		t.Fatalf("duplicate: %v", err)
	}
}

func TestRegistrationValidation(t *testing.T) {
	for _, input := range [][3]string{{"", "a@example.com", "password123"}, {"school", "bad", "password123"}, {"school", "Name <a@example.com>", "password123"}, {"school", "a@example.com", "short"}, {strings.Repeat("a", 51), "a@example.com", "password123"}, {"school", "a@example.com", strings.Repeat("a", 1025)}} {
		store := &storeStub{}
		_, err := (Registration{Store: store}).Register(context.Background(), input[0], input[1], input[2])
		if !errors.Is(err, ErrInvalidRegistration) || store.called {
			t.Fatalf("invalid input accepted: %v", input[:2])
		}
	}
}
