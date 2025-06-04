#!/bin/sh

echo "🚀 Starting GoSocial Test Backend..."

# Wait for database to be ready
echo "⏳ Waiting for database..."
while ! migrate -path=./cmd/migrate/migrations -database="$DB_ADDR" version 2>/dev/null; do
    echo "Database not ready, waiting..."
    sleep 2
done

echo "🔄 Running database migrations..."
if migrate -path=./cmd/migrate/migrations -database="$DB_ADDR" up; then
    echo "✅ Migrations completed successfully!"
else
    echo "❌ Migration failed!"
    exit 1
fi

echo "🏃 Starting GoSocial application..."
exec ./bin/main