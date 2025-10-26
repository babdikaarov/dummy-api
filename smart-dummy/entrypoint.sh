#!/bin/bash

set -e

echo "========================================="
echo "  Application Startup Script"
echo "========================================="

# Wait a bit for database to be fully ready
echo "⏳ Waiting for database connection..."
sleep 5

# Generate database schema
echo "📋 Generating database schema..."
npm run db:generate

# Run migrations
echo "🔄 Running database migrations..."
npm run db:migrate

# Seed database
echo "🌱 Seeding database..."
npm run db:seed

echo "========================================="
echo "✅ Database setup completed!"
echo "========================================="

# Start the application
echo "🚀 Starting application..."
exec npm run start:dev
