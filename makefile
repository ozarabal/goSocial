include .env
MIGRATIONS_PATH = ./cmd/migrate/migrations

.PHONY : migrate-create
migration:
	@migrate create -seq -ext sql -dir $(MIGRATIONS_PATH) $(filter-out $@,$(MAKECMDGOALS))

.PHONY: migrate-up
migrate-up:
	@docker exec -it app migrate -path=$(MIGRATIONS_PATH) -database=$(DB_ADDR) up || $(MAKE) fix-dirty

.PHONY: migrate-down
migrate-down:
	@docker exec -it app migrate -path=$(MIGRATIONS_PATH) -database=$(DB_ADDR) down $(filter-out $@,$(MAKECMDGOALS)) || $(MAKE) fix-dirty

.PHONY: check-dirty
check-dirty:
	@docker exec -it db psql $(DB_ADDR) -c "SELECT version, dirty FROM schema_migrations ORDER BY version DESC LIMIT 1;" | grep -q 't' && echo "Dirty database detected!" || echo "Database is clean."

.PHONY: fix-dirty
fix-dirty:
	@echo "Fixing dirty database..."
	@docker exec -it db psql $(DB_ADDR) -c "UPDATE schema_migrations SET dirty = FALSE WHERE dirty = TRUE;"
	@echo "Dirty database flag cleared. You can now reapply the migration."



.PHONY: seed
seed:
	@docker exec -it app go run cmd/migrate/seed/main.go

.PHONY: gen-docs
gen-docs:
	@swag init -g ./api/main.go -d cmd,internal && swag fmt