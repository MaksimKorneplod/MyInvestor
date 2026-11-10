package repository

import (
	"context"
	"errors"
	"github.com/MaksimKorneplod/MyInvestor/internal/service"
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
