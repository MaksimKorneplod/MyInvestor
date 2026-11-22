package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"
)

var (
	ErrAccountNotFound    = errors.New("account not found")
	ErrInvalidCredentials = errors.New("invalid credentials")
)

type SchoolCredentials struct {
	ID   string `json:"school_id"`
	Hash string `json:"hash"`
}

type LoginStore interface {
	FindSchool(ctx context.Context, email string) (SchoolCredentials, error)
	CreateSchoolSession(ctx context.Context, schoolID, tokenHash string, expiresAt time.Time) error
}

type Login struct{ Store LoginStore }

const sessionLifetime = 24 * time.Hour

const dummyPasswordHash = "pbkdf2-sha256$600000$AAECAwQFBgcICQoLDA0ODw$GDXXLNrdj58h6eV5zf2eM0Cq4zdD1nrEcSgwQq0vwIU"

type LoginSession struct {
	AccessToken string `json:"accessToken"`
	TokenType   string `json:"tokenType"`
	ExpiresIn   int64  `json:"expiresIn"`
}

func (s Login) Authenticate(ctx context.Context, email, password string) (LoginSession, error) {
	if s.Store == nil {
		return LoginSession{}, errors.New("login store is nil")
	}
	email = strings.ToLower(strings.TrimSpace(email))
	school, err := s.Store.FindSchool(ctx, email)
	if err != nil {
		if errors.Is(err, ErrAccountNotFound) {
			if verifyErr := VerifyPassword(password, dummyPasswordHash); verifyErr != nil && !errors.Is(verifyErr, ErrPasswordMismatch) {
				return LoginSession{}, fmt.Errorf("verify dummy password hash: %w", verifyErr)
			}
			return LoginSession{}, ErrInvalidCredentials
		}
		return LoginSession{}, fmt.Errorf("find school: %w", err)
	}
	if err := VerifyPassword(password, school.Hash); err != nil {
		if errors.Is(err, ErrPasswordMismatch) {
			return LoginSession{}, ErrInvalidCredentials
		}
		return LoginSession{}, fmt.Errorf("verify stored password: %w", err)
	}

	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return LoginSession{}, fmt.Errorf("generate session token: %w", err)
	}
	token := base64.RawURLEncoding.EncodeToString(tokenBytes)
	tokenHash := sha256.Sum256([]byte(token))
	expiresAt := time.Now().Add(sessionLifetime)
	if err := s.Store.CreateSchoolSession(ctx, school.ID, hex.EncodeToString(tokenHash[:]), expiresAt); err != nil {
		return LoginSession{}, fmt.Errorf("create school session: %w", err)
	}

	return LoginSession{
		AccessToken: token,
		TokenType:   "Bearer",
		ExpiresIn:   int64(sessionLifetime.Seconds()),
	}, nil
}
