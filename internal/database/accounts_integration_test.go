package database

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"
)

func TestSchoolAccountsMigration(t *testing.T) {
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("set TEST_DATABASE_URL to run PostgreSQL migration tests")
	}
	parsed, err := url.Parse(databaseURL)
	if err != nil || !strings.HasSuffix(parsed.Path, "_test") {
		t.Fatal("TEST_DATABASE_URL must target an isolated *_test database")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	conn, err := pgx.Connect(ctx, databaseURL)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close(context.Background())
	schema := fmt.Sprintf("accounts_test_%d", time.Now().UnixNano())
	if _, err = conn.Exec(ctx, "CREATE SCHEMA "+schema); err != nil {
		t.Fatal(err)
	}
	defer conn.Exec(context.Background(), "DROP SCHEMA "+schema+" CASCADE")
	if _, err = conn.Exec(ctx, "SET search_path TO "+schema); err != nil {
		t.Fatal(err)
	}
	up, err := os.ReadFile("../../migrations/000002_school_accounts.up.sql")
	if err != nil {
		t.Fatal(err)
	}
	down, err := os.ReadFile("../../migrations/000002_school_accounts.down.sql")
	if err != nil {
		t.Fatal(err)
	}
	for cycle := 0; cycle < 2; cycle++ {
		if _, err = conn.Exec(ctx, string(up)); err != nil {
			t.Fatal(err)
		}
		var school, student string
		if err = conn.QueryRow(ctx, `INSERT INTO school(login,email,password_hash) VALUES ('school','school@example.com','hash') RETURNING school_id::text`).Scan(&school); err != nil {
			t.Fatal(err)
		}
		if _, err = conn.Exec(ctx, `INSERT INTO school(login,email,password_hash) VALUES ('other','SCHOOL@example.com','hash')`); err == nil {
			t.Fatal("duplicate email accepted")
		}
		if err = conn.QueryRow(ctx, `INSERT INTO student(school_id,login,email,password_hash,expiration_date) VALUES ($1,'student','student@example.com','hash',CURRENT_DATE+365) RETURNING student_id::text`, school).Scan(&student); err != nil {
			t.Fatal(err)
		}
		var balance int64
		var streak, favorites int
		if err = conn.QueryRow(ctx, `SELECT balance,streak_count FROM student_info WHERE student_id=$1`, student).Scan(&balance, &streak); err != nil {
			t.Fatal(err)
		}
		if balance != 100000 || streak != 0 {
			t.Fatalf("profile defaults: %d %d", balance, streak)
		}
		if err = conn.QueryRow(ctx, `SELECT count(*) FROM favorite_stocks WHERE student_id=$1`, student).Scan(&favorites); err != nil || favorites != 0 {
			t.Fatal("favorites must start empty", err)
		}
		if _, err = conn.Exec(ctx, `INSERT INTO student_info(student_id) VALUES ($1)`, student); err == nil {
			t.Fatal("duplicate profile accepted")
		}
		if _, err = conn.Exec(ctx, `INSERT INTO favorite_stocks VALUES (gen_random_uuid(),'TEST')`); err == nil {
			t.Fatal("orphan favorite accepted")
		}
		if _, err = conn.Exec(ctx, `UPDATE student_info SET balance=-1 WHERE student_id=$1`, student); err == nil {
			t.Fatal("negative balance accepted")
		}
		if _, err = conn.Exec(ctx, string(down)); err != nil {
			t.Fatal(err)
		}
	}
}
