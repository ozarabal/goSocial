include .env
MIGRATIONS_PATH = ./cmd/migrate/migrations

.PHONY: test
test:
	@go test -v ./...

.PHONY : migrate-create
migration:
	@migrate create -seq -ext sql -dir $(MIGRATIONS_PATH) $(filter-out $@,$(MAKECMDGOALS))

.PHONY: migrate-up
migrate-up:
	@migrate -path=$(MIGRATIONS_PATH) -database=postgres://postgres:wHidiAtqVlylthCbmWDvxkWQklfchquR@postgres.railway.internal:5432/railway
 up || $(MAKE) fix-dirty

.PHONY: migrate-down
migrate-down:
	@migrate -path=$(MIGRATIONS_PATH) -database=$(DB_ADDR) down $(filter-out $@,$(MAKECMDGOALS)) || $(MAKE) fix-dirty

.PHONY: check-dirty
check-dirty:
	@railway run psql $(DB_ADDR) -c "SELECT version, dirty FROM schema_migrations ORDER BY version DESC LIMIT 1;" | grep -q 't' && echo "Dirty database detected!" || echo "Database is clean."

.PHONY: fix-dirty
fix-dirty:
	@echo "Fixing dirty database..."
	@railway run psql $(DB_ADDR) -c "UPDATE schema_migrations SET dirty = FALSE WHERE dirty = TRUE;"
	@echo "Dirty database flag cleared. You can now reapply the migration."



.PHONY: seed
seed:
	@docker exec -it app go run cmd/migrate/seed/main.go

.PHONY: gen-docs
gen-docs:
	@swag init -g ./api/main.go -d cmd,internal && swag fmt