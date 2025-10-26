#!/bin/bash

set -e

echo "========================================="
echo "  Production Startup Script"
echo "========================================="

# Wait a bit for database to be fully ready
echo "⏳ Waiting for database connection..."
sleep 10

# Generate database schema
echo "📋 Generating database schema..."
npm run db:generate

# Run migrations
echo "🔄 Running database migrations..."
npm run db:migrate

# Seed database
echo "🌱 Seeding database..."
node dist/src/database/seed.js

echo "========================================="
echo "✅ Database setup completed!"
echo "========================================="

# Start the application in production mode
echo "🚀 Starting production application..."
exec node dist/src/main.js
