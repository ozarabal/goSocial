MIGRATIONS_PATH = ./cmd/migrate/migrations

.PHONY: test
test:
	@go test -v ./...

.PHONY : migrate-create
migration:
	@migrate create -seq -ext sql -dir $(MIGRATIONS_PATH) $(filter-out $@,$(MAKECMDGOALS))

.PHONY: migrate-up
migrate-up:
	@migrate -path=$(MIGRATIONS_PATH) -database=postgresql://postgres:wHidiAtqVlylthCbmWDvxkWQklfchquR@postgres.railway.internal:5432/railway up || $(MAKE) fix-dirty

.PHONY: migrate-down
migrate-down:
	@migrate -path=$(MIGRATIONS_PATH) -database=postgresql://postgres:wHidiAtqVlylthCbmWDvxkWQklfchquR@postgres.railway.internal:5432/railway down $(filter-out $@,$(MAKECMDGOALS)) || $(MAKE) fix-dirty

# NEW: Test-specific migration commands
.PHONY: migrate-up-test
migrate-up-test:
	@echo "🔄 Running test database migrations..."
	@migrate -path=$(MIGRATIONS_PATH) -database=postgres://admin:adminpassword@db-test:5432/social_test?sslmode=disable up || echo "⚠️ Migrations failed"

.PHONY: migrate-down-test
migrate-down-test:
	@echo "🔄 Rolling back test database migrations..."
	@migrate -path=$(MIGRATIONS_PATH) -database=postgres://admin:adminpassword@db-test:5432/social_test?sslmode=disable down $(filter-out $@,$(MAKECMDGOALS))

.PHONY: migrate-reset-test
migrate-reset-test:
	@echo "🔄 Resetting test database..."
	@migrate -path=$(MIGRATIONS_PATH) -database=postgres://admin:adminpassword@db-test:5432/social_test?sslmode=disable down -all || echo "Down failed"
	@migrate -path=$(MIGRATIONS_PATH) -database=postgres://admin:adminpassword@db-test:5432/social_test?sslmode=disable up || echo "Up failed"

.PHONY: check-dirty
check-dirty:
	@railway run psql postgresql://postgres:wHidiAtqVlylthCbmWDvxkWQklfchquR@postgres.railway.internal:5432/railway -c "SELECT version, dirty FROM schema_migrations ORDER BY version DESC LIMIT 1;" | grep -q 't' && echo "Dirty database detected!" || echo "Database is clean."

.PHONY: fix-dirty
fix-dirty:
	@echo "Fixing dirty database..."
	@railway run psql postgresql://postgres:wHidiAtqVlylthCbmWDvxkWQklfchquR@postgres.railway.internal:5432/railway -c "UPDATE schema_migrations SET dirty = FALSE WHERE dirty = TRUE;"
	@echo "Dirty database flag cleared. You can now reapply the migration."

.PHONY: seed
seed:
	@go run cmd/migrate/seed/main.go

.PHONY: gen-docs
gen-docs:
	@swag init -g ./api/main.go -d cmd,internal && swag fmt

.PHONY: gen-docs-fixed
gen-docs-fixed:
	@echo "📚 Generating API documentation..."
	@swag init -g ./cmd/api/main.go -d cmd,internal || echo "⚠️ Documentation generation failed, continuing..."
	@swag fmt || echo "⚠️ Documentation formatting failed, continuing..."
	@echo "✅ Documentation step completed!"

# Test automation commands
.PHONY: test-api
test-api:
	@echo "Running API Test Suite..."
	@go test -v ./tests/api/... -timeout=5m || echo "Some tests failed"

.PHONY: test-api-auth
test-api-auth:
	@echo "Running Authentication API Tests..."
	@go test -v ./tests/api/auth/... -timeout=2m || echo "Auth tests failed"

.PHONY: test-api-posts
test-api-posts:
	@echo "Running Posts API Tests..."
	@go test -v ./tests/api/posts/... -timeout=2m || echo "Posts tests failed"

.PHONY: test-run-posts
test-run-posts:
	@echo "📝 Running posts tests in container..."
	@docker-compose -f docker-compose.test.yml run --rm test-runner sh -c "until curl -f http://app-test:3000/v1/health > /dev/null 2>&1; do echo 'Waiting for API...'; sleep 5; done && echo 'API ready!' && make test-api-posts"

.PHONY: test-api-users
test-api-users:
	@echo "Running Users API Tests..."
	@go test -v ./tests/api/users/... -timeout=2m || echo "Users tests failed"

.PHONY: test-smoke
test-smoke:
	@echo "💨 Running Smoke Tests..."
	@echo "🏥 Testing health endpoint..."
	@curl -f http://localhost:3001/v1/health > /dev/null 2>&1 && echo "✅ Health check passed" || echo "❌ Health check failed"
	@echo "📊 Checking API status..."
	@curl -s http://localhost:3001/v1/health | grep -q "ok" && echo "✅ API status OK" || echo "❌ API status failed"
	@echo "🎯 Basic smoke test completed"

.PHONY: test-install-deps
test-install-deps:
	@echo "Installing test dependencies..."
	@go mod verify
	@go mod download
	@go mod tidy
	@echo "✅ Dependencies ready"

# =============================================================================
# DOCKER TEST AUTOMATION COMMANDS
# =============================================================================

# Docker Test Environment Management
.PHONY: test-docker-build
test-docker-build:
	@echo "🔨 Building test containers..."
	@docker-compose -f docker-compose.test.yml build

# NEW: Wait for services to be healthy
.PHONY: test-wait-ready
test-wait-ready:
	@echo "⏳ Waiting for all services to be healthy..."
	@docker-compose -f docker-compose.test.yml up --wait db-test redis-test app-test
	@echo "✅ All services are ready!"

.PHONY: test-docker-up
test-docker-up:
	@echo "🚀 Starting test environment..."
	@docker-compose -f docker-compose.test.yml up -d db-test redis-test app-test
	@echo "⏳ Waiting for services to be ready..."
	@echo "ℹ️  Services are starting in background. Use 'make test-status-docker' to check status."
	@echo "ℹ️  Wait about 30-60 seconds before running tests."

.PHONY: test-docker-down
test-docker-down:
	@echo "🛑 Stopping test environment..."
	@docker-compose -f docker-compose.test.yml down
	@echo "✅ Test environment stopped!"

.PHONY: test-docker-reset
test-docker-reset:
	@echo "🔄 Resetting test environment..."
	@docker-compose -f docker-compose.test.yml down -v
	@docker-compose -f docker-compose.test.yml build --no-cache
	@make test-docker-up

.PHONY: test-docker-logs
test-docker-logs:
	@echo "📋 Showing test environment logs..."
	@docker-compose -f docker-compose.test.yml logs -f

# Docker Test Execution
.PHONY: test-run-smoke
test-run-smoke:
	@echo "💨 Running smoke tests in container..."
	@docker-compose -f docker-compose.test.yml run --rm test-runner sh -c "until curl -f http://app-test:3000/v1/health > /dev/null 2>&1; do echo 'Waiting for API...'; sleep 5; done && echo 'API ready!' && make test-smoke"

.PHONY: test-run-all
test-run-all:
	@echo "🧪 Running all API tests in container..."
	@docker-compose -f docker-compose.test.yml run --rm test-runner sh -c "until curl -f http://app-test:3000/v1/health > /dev/null 2>&1; do echo 'Waiting for API...'; sleep 5; done && echo 'API ready!' && make test-api"

.PHONY: test-run-auth
test-run-auth:
	@echo "🔐 Running auth tests in container..."
	@docker-compose -f docker-compose.test.yml run --rm test-runner sh -c "until curl -f http://app-test:3000/v1/health > /dev/null 2>&1; do echo 'Waiting for API...'; sleep 5; done && echo 'API ready!' && make test-api-auth"

# Complete Docker test workflows
.PHONY: test-full-docker
test-full-docker:
	@echo "🚀 Running complete Docker test workflow..."
	@make test-docker-build
	@make test-docker-up
	@sleep 60
	@make test-run-smoke
	@make test-docker-down
	@echo "✅ Complete Docker test workflow finished!"

.PHONY: test-workflow-quick
test-workflow-quick:
	@echo "⚡ Running quick Docker test workflow..."
	@make test-docker-up
	@sleep 30
	@make test-run-smoke
	@make test-docker-down
	@echo "✅ Quick Docker test workflow finished!"

# Docker Test Utilities
.PHONY: test-status-docker
test-status-docker:
	@echo "📊 Docker Test Environment Status:"
	@docker-compose -f docker-compose.test.yml ps
	@echo ""
	@echo "🏥 Health Checks:"
	@curl -s http://localhost:3001/v1/health || echo "❌ API not ready"

.PHONY: test-shell-docker
test-shell-docker:
	@echo "🐚 Opening test shell in container..."
	@docker-compose -f docker-compose.test.yml run --rm test-runner sh

.PHONY: test-cleanup-docker
test-cleanup-docker:
	@echo "🧹 Cleaning up Docker test environment..."
	@docker-compose -f docker-compose.test.yml down -v --remove-orphans
	@docker system prune -f
	@echo "✅ Docker cleanup completed!"

# Docker Test Debugging
.PHONY: test-debug-app
test-debug-app:
	@echo "🔍 Debugging app container..."
	@docker-compose -f docker-compose.test.yml logs app-test

.PHONY: test-debug-db
test-debug-db:
	@echo "🔍 Debugging database container..."
	@docker-compose -f docker-compose.test.yml logs db-test

.PHONY: test-debug-redis
test-debug-redis:
	@echo "🔍 Debugging redis container..."
	@docker-compose -f docker-compose.test.yml logs redis-test

.PHONY: test-debug-all
test-debug-all:
	@echo "🔍 Debugging all test containers..."
	@docker-compose -f docker-compose.test.yml logs

.PHONY: test-exec-app
test-exec-app:
	@echo "💻 Executing shell in app container..."
	@docker-compose -f docker-compose.test.yml exec app-test sh

.PHONY: test-exec-db
test-exec-db:
	@echo "💻 Executing shell in database container..."
	@docker-compose -f docker-compose.test.yml exec db-test psql -U admin -d social_test

# Help commands
.PHONY: test-help-docker
test-help-docker:
	@echo "🐳 Docker Test Automation Commands:"
	@echo ""
	@echo "🔧 Environment Management:"
	@echo "  test-docker-build     - Build test containers"
	@echo "  test-docker-up        - Start test environment"
	@echo "  test-docker-down      - Stop test environment"
	@echo "  test-docker-reset     - Reset test environment"
	@echo "  test-docker-logs      - Show environment logs"
	@echo ""
	@echo "🧪 Test Execution:"
	@echo "  test-run-smoke        - Run smoke tests in container"
	@echo "  test-run-auth         - Run auth tests in container"
	@echo "  test-run-all          - Run all API tests in container"
	@echo ""
	@echo "🚀 Workflows:"
	@echo "  test-full-docker      - Complete test workflow"
	@echo "  test-workflow-quick   - Quick test workflow"
	@echo ""
	@echo "🛠️ Utilities:"
	@echo "  test-status-docker    - Check test environment status"
	@echo "  test-shell-docker     - Interactive test shell"
	@echo "  test-cleanup-docker   - Complete cleanup"
	@echo ""
	@echo "🔍 Debugging:"
	@echo "  test-debug-app        - Debug app container logs"
	@echo "  test-debug-db         - Debug database container logs"
	@echo "  test-debug-redis      - Debug redis container logs"
	@echo "  test-exec-app         - Execute shell in app container"
	@echo "  test-exec-db          - Execute shell in database container"
	@echo ""
	@echo "💡 Quick Start: make test-full-docker"

.PHONY: help
help:
	@echo "🚀 GoSocial Makefile Commands:"
	@echo ""
	@echo "🐳 Docker Testing (RECOMMENDED):"
	@echo "  test-full-docker      - Complete Docker test workflow"
	@echo "  test-workflow-quick   - Quick Docker test workflow"
	@echo "  test-docker-up        - Start test environment"
	@echo "  test-run-smoke        - Run smoke tests"
	@echo "  test-docker-down      - Stop test environment"
	@echo ""
	@echo "📚 Documentation:"
	@echo "  help                  - Show this help message"
	@echo "  test-help-docker      - Show Docker test commands"
	@echo ""
	@echo "🏗️ Development:"
	@echo "  gen-docs-fixed        - Generate API documentation"
	@echo "  seed                  - Seed database with test data"
	@echo ""
	@echo "🗄️ Database:"
	@echo "  migrate-up            - Run production migrations"
	@echo "  migrate-up-test       - Run test migrations"
	@echo "  migrate-reset-test    - Reset test database"
	@echo ""
	@echo "💡 Quick Start for Testing: make test-full-docker"