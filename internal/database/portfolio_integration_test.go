package database

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

func TestPortfolioMigration(t *testing.T) {
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

	exec := func(query string, args ...any) {
		t.Helper()
		if _, err := conn.Exec(ctx, query, args...); err != nil {
			t.Fatal(err)
		}
	}
	reject := func(code, query string, args ...any) {
		t.Helper()
		_, err := conn.Exec(ctx, query, args...)
		var pgErr *pgconn.PgError
		if !errors.As(err, &pgErr) || pgErr.Code != code {
			t.Fatalf("want PostgreSQL error %s, got %v", code, err)
		}
	}
	migrate := func(name string) {
		t.Helper()
		sql, err := os.ReadFile("../../migrations/" + name + ".sql")
		if err != nil {
			t.Fatal(err)
		}
		exec(string(sql))
	}

	schema := fmt.Sprintf("portfolio_test_%d", time.Now().UnixNano())
	exec("CREATE SCHEMA " + schema)
	defer conn.Exec(context.Background(), "DROP SCHEMA "+schema+" CASCADE")
	exec("SET search_path TO " + schema)
	migrate("000001_init_schema.up")
	migrate("000002_school_accounts.up")
	migrate("000004_portfolio.up")

	exec("INSERT INTO school (login, email, password_hash) VALUES ('school', 'school@example.org', 'hash')")
	var student string
	err = conn.QueryRow(ctx, `INSERT INTO student (school_id, login, email, password_hash, expiration_date)
		SELECT school_id, 'student', 'student@example.org', 'hash', CURRENT_DATE + 1 FROM school
		RETURNING student_id::text`).Scan(&student)
	if err != nil {
		t.Fatal(err)
	}
	const isin = "RU0009029540"
	const instrument = "INSERT INTO trading_instrument (isin) VALUES ($1)"
	exec(instrument, isin)
	reject("23505", instrument, isin)
	reject("23514", instrument, "ru0009029540")
	reject("23514", instrument, "SBER")

	const position = "INSERT INTO portfolio_position (student_id, instrument_isin, quantity, average_price) VALUES ($1, $2, $3, 10.12345678)"
	exec(position, student, isin, 2)
	reject("23505", position, student, isin, 2)
	reject("23503", position, student, "RU0007661625", 2)
	reject("23503", position, "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa", isin, 2)
	reject("23514", "UPDATE portfolio_position SET quantity = 0")
	reject("23514", "UPDATE portfolio_position SET average_price = 0")
	reject("23514", "UPDATE portfolio_position SET average_price = 'NaN'")

	const trade = "INSERT INTO trade_operation (student_id, instrument_isin, side, quantity, price) VALUES ($1, $2, $3, $4, $5)"
	exec(trade, student, isin, "buy", 2, "10.12000000")
	exec(trade, student, isin, "sell", 2, "10.12345678")
	reject("23514", trade, student, isin, "hold", 2, "10.12")
	reject("23514", trade, student, isin, "sell", -1, "10.12")
	reject("23514", trade, student, isin, "sell", 0, "10.12")
	reject("23514", trade, student, isin, "sell", 1, "-10.12")
	reject("23514", trade, student, isin, "sell", 1, "NaN")
	reject("23503", trade, student, "RU0007661625", "buy", 1, "10.12")
	reject("23503", trade, "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa", isin, "buy", 1, "10.12")
	reject("23503", "DELETE FROM trading_instrument WHERE isin = $1", isin)

	exec("DELETE FROM portfolio_position WHERE student_id = $1 AND instrument_isin = $2", student, isin)
	var count int
	if err := conn.QueryRow(ctx, "SELECT count(*) FROM trade_operation WHERE executed_at IS NOT NULL").Scan(&count); err != nil || count != 2 {
		t.Fatalf("closing the position must preserve both timestamped trades: count=%d, err=%v", count, err)
	}
	var price string
	if err := conn.QueryRow(ctx, "SELECT price::text FROM trade_operation WHERE side = 'sell'").Scan(&price); err != nil || price != "10.12345678" {
		t.Fatalf("trade price must preserve decimal precision: price=%q, err=%v", price, err)
	}

	migrate("000004_portfolio.down")
	reject("42P01", "SELECT 1 FROM portfolio_position")
	reject("42P01", "SELECT 1 FROM trade_operation")
	reject("42P01", "SELECT 1 FROM trading_instrument")
	migrate("000004_portfolio.up")
	exec(instrument, isin)
	exec(position, student, isin, 1)
}
