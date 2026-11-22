package repository

import (
	"context"
	"errors"
	"time"

	"github.com/MaksimKorneplod/MyInvestor/internal/service"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type SchoolStore struct{ Pool *pgxpool.Pool }

func (s SchoolStore) CreateSchool(ctx context.Context, login, email, hash string) (service.School, error) {
	var school service.School
	err := s.Pool.QueryRow(ctx, `INSERT INTO school(login,email,password_hash) VALUES ($1,$2,$3) RETURNING school_id::text,login,email`, login, email, hash).Scan(&school.ID, &school.Login, &school.Email)
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return service.School{}, service.ErrAccountExists
	}
	return school, err
}

func (s SchoolStore) FindSchool(ctx context.Context, email string) (service.SchoolCredentials, error) {
	var school service.SchoolCredentials
	err := s.Pool.QueryRow(ctx, `SELECT school_id::text,password_hash FROM school WHERE lower(email)=lower($1)`, email).Scan(&school.ID, &school.Hash)
	if errors.Is(err, pgx.ErrNoRows) {
		return service.SchoolCredentials{}, service.ErrAccountNotFound
	}
	return school, err
}

func (s SchoolStore) CreateSchoolSession(ctx context.Context, schoolID, tokenHash string, expiresAt time.Time) error {
	_, err := s.Pool.Exec(ctx, `
		INSERT INTO school_sessions (school_id, token_hash, expires_at)
		VALUES ($1, $2, $3)`,
		schoolID, tokenHash, expiresAt,
	)
	return err
}
