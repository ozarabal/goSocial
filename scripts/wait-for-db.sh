#!/bin/sh
echo "🚀 Starting GoSocial Test Backend..."
echo "⏳ Waiting for database..."

# Wait for database to be ready
until pg_isready -h db-test -p 5432 -U admin; do
  echo "Database not ready, waiting..."
  sleep 2
done

echo "✅ Database is ready!"

# Test database connection
echo "🔍 Testing database connection..."
PGPASSWORD=adminpassword psql -h db-test -U admin -d social_test -c "SELECT 1;" > /dev/null 2>&1
if [ $? -eq 0 ]; then
    echo "✅ Database connection successful!"
else
    echo "❌ Database connection failed!"
    exit 1
fi

echo "🔄 Running migrations..."
migrate -path=./cmd/migrate/migrations -database=postgres://admin:adminpassword@db-test:5432/social_test?sslmode=disable up || echo "⚠️ Migration failed, continuing anyway"

echo "🚀 Starting application..."
exec ./bin/main