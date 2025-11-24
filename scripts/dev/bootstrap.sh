#!/bin/bash

echo "Starting all services..."
docker-compose up -d

echo "Waiting for PostgreSQL to be ready..."
until docker-compose exec postgres_write pg_isready -U postgres -d write_model > /dev/null 2>&1; do
  sleep 1
done

echo "Running migrations..."
go run scripts/migrate/migrate.go

echo "Seeding data..."
make seed

echo "Deploying SAM application..."
make deploy

echo "Bootstrap complete."
