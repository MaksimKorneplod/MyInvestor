package service

import (
	"context"
	"crypto/pbkdf2"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"net/mail"
	"strings"
	"unicode/utf8"
)

var ErrInvalidRegistration = errors.New("invalid registration")
var ErrAccountExists = errors.New("email or login already registered")

type School struct {
	ID    string `json:"school_id"`
	Login string `json:"login"`
	Email string `json:"email"`
}

type SchoolStore interface {
	CreateSchool(context.Context, string, string, string) (School, error)
}

type Registration struct{ Store SchoolStore }

func (s Registration) Register(ctx context.Context, login, email, password string) (School, error) {
	login = strings.TrimSpace(login)
	email = strings.ToLower(strings.TrimSpace(email))
	address, err := mail.ParseAddress(email)
	validLogin := login != "" && utf8.RuneCountInString(login) <= 50 && !strings.ContainsRune(login, '\x00')
	validEmail := len(email) <= 254 && err == nil && address.Address == email && strings.Contains(email, "@")
	validPassword := len(password) >= 8 && len(password) <= 1024 && strings.TrimSpace(password) != ""
	if !validLogin || !validEmail || !validPassword {
		return School{}, fmt.Errorf("%w: login (1–50 characters), valid email and password (8–1024 bytes) are required", ErrInvalidRegistration)
	}
	hash, err := HashPassword(password)
	if err != nil {
		return School{}, err
	}
	return s.Store.CreateSchool(ctx, login, email, hash)
}

// HashPassword uses PBKDF2-HMAC-SHA256 with a fresh 128-bit salt.
// The algorithm, work factor, salt and derived key are stored for future login verification.
func HashPassword(password string) (string, error) {
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}
	key, err := pbkdf2.Key(sha256.New, password, salt, 600000, 32)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("pbkdf2-sha256$600000$%s$%s", base64.RawStdEncoding.EncodeToString(salt), base64.RawStdEncoding.EncodeToString(key)), nil
}
