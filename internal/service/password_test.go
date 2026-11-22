package service

import (
	"errors"
	"testing"
)

func TestVerifyPassword(t *testing.T) {
	hash, err := HashPassword("correct horse battery staple")
	if err != nil {
		t.Fatal(err)
	}

	if err := VerifyPassword("correct horse battery staple", hash); err != nil {
		t.Fatalf("VerifyPassword() error = %v", err)
	}
	if err := VerifyPassword("wrong password", hash); err == nil {
		t.Fatal("VerifyPassword() succeeded for wrong password")
	}
}

func TestVerifyPasswordRejectsMalformedHash(t *testing.T) {
	if err := VerifyPassword("password", "not-a-password-hash"); !errors.Is(err, ErrInvalidPasswordHash) {
		t.Fatalf("VerifyPassword() error = %v, want %v", err, ErrInvalidPasswordHash)
	}
}
