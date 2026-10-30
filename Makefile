MIGRATIONS_DIR := ./migrations
DATABASE_URL ?= $(shell sed -n 's/^DATABASE_URL=//p' .env 2>/dev/null | head -n 1)

.PHONY: migrate-up migrate-down migrate-version test-db-up test-db-down test-integration

migrate-up:
	@test -n "$(DATABASE_URL)" || (echo "Set DATABASE_URL before applying migrations" >&2; exit 1)
	@migrate -path "$(MIGRATIONS_DIR)" -database "$(DATABASE_URL)" up

migrate-down:
	@test -n "$(DATABASE_URL)" || (echo "Set DATABASE_URL before rolling back migrations" >&2; exit 1)
	@migrate -path "$(MIGRATIONS_DIR)" -database "$(DATABASE_URL)" down 1

migrate-version:
	@test -n "$(DATABASE_URL)" || (echo "Set DATABASE_URL before checking migration status" >&2; exit 1)
	@migrate -path "$(MIGRATIONS_DIR)" -database "$(DATABASE_URL)" version

test-db-up:
	docker compose --profile test up -d --wait db-test

test-db-down:
	docker compose stop db-test
	docker compose rm -f db-test

test-integration:
	@test -n "$(TEST_DATABASE_URL)" || (echo "Set TEST_DATABASE_URL to an isolated *_test database" >&2; exit 1)
	go test -count=1 -run '^TestConnectAgainstIsolatedDatabase$$' ./internal/database
