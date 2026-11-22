package service

import (
	"context"
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
