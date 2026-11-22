package service

import (
	"crypto/pbkdf2"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strconv"
	"strings"
)

const (
	passwordHashAlgorithm  = "pbkdf2-sha256"
	passwordHashIterations = 600000
	passwordHashKeyLength  = 32
	passwordSaltLength     = 16
)

var ErrInvalidPasswordHash = errors.New("invalid password hash")
var ErrPasswordMismatch = errors.New("password does not match")

// HashPassword uses PBKDF2-HMAC-SHA256 with a fresh 128-bit salt.
// The algorithm, work factor, salt and derived key are stored for future login verification.
func HashPassword(password string) (string, error) {
	salt := make([]byte, passwordSaltLength)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}
	key, err := pbkdf2.Key(sha256.New, password, salt, passwordHashIterations, passwordHashKeyLength)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s$%d$%s$%s",
		passwordHashAlgorithm,
		passwordHashIterations,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(key),
	), nil
}

func VerifyPassword(password, hash string) error {
	parts := strings.Split(hash, "$")
	if len(parts) != 4 || parts[0] != passwordHashAlgorithm {
		return ErrInvalidPasswordHash
	}
	iterations, err := strconv.Atoi(parts[1])
	if err != nil {
		return fmt.Errorf("%w: invalid iteration count", ErrInvalidPasswordHash)
	}
	if iterations != passwordHashIterations {
		return fmt.Errorf("%w: unsupported iteration count", ErrInvalidPasswordHash)
	}
	salt, err := base64.RawStdEncoding.DecodeString(parts[2])
	if err != nil {
		return fmt.Errorf("%w: invalid salt encoding", ErrInvalidPasswordHash)
	}
	if len(salt) != passwordSaltLength {
		return fmt.Errorf("%w: invalid salt length", ErrInvalidPasswordHash)
	}
	expectedKey, err := base64.RawStdEncoding.DecodeString(parts[3])
	if err != nil {
		return fmt.Errorf("%w: invalid key encoding", ErrInvalidPasswordHash)
	}
	if len(expectedKey) != passwordHashKeyLength {
		return fmt.Errorf("%w: invalid key length", ErrInvalidPasswordHash)
	}
	actualKey, err := pbkdf2.Key(sha256.New, password, salt, iterations, passwordHashKeyLength)
	if err != nil {
		return fmt.Errorf("derive password key: %w", err)
	}
	if subtle.ConstantTimeCompare(actualKey, expectedKey) != 1 {
		return ErrPasswordMismatch
	}
	return nil
}
