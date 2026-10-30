package database

import (
	"context"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"
)

func TestConnectAgainstIsolatedDatabase(t *testing.T) {
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("set TEST_DATABASE_URL to run the PostgreSQL integration test")
	}

	parsedURL, err := url.Parse(databaseURL)
	if err != nil {
		t.Fatalf("parse TEST_DATABASE_URL: %v", err)
	}
	if !strings.HasSuffix(strings.TrimPrefix(parsedURL.Path, "/"), "_test") {
		t.Fatal("TEST_DATABASE_URL must point to a database whose name ends with _test")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pool, err := Connect(ctx, databaseURL)
	if err != nil {
		t.Fatalf("connect to test database: %v", err)
	}
	t.Cleanup(pool.Close)
}
