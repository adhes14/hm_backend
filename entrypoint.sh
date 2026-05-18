#!/bin/sh
set -e

# Run migrations if RUN_MIGRATIONS is not set to "false"
if [ "$RUN_MIGRATIONS" != "false" ]; then
  if [ -n "$DATABASE_URL" ]; then
    echo "========================================="
    echo "Running database migrations via DATABASE_URL..."
    echo "========================================="
    # golang-migrate accepts standard connection strings
    /usr/local/bin/migrate -path /app/migrations -database "$DATABASE_URL" up || {
      echo "Migration failed! Exiting."
      exit 1
    }
    echo "Migrations applied successfully!"
  elif [ -n "$POSTGRES_HOST" ]; then
    echo "========================================="
    echo "Running database migrations via POSTGRES_HOST..."
    echo "========================================="
    CONN_STR="postgres://${POSTGRES_USER:-hospital}:${POSTGRES_PASSWORD:-hospital_pass}@${POSTGRES_HOST}:${POSTGRES_PORT:-5432}/${POSTGRES_DB:-hospital_db}?sslmode=disable"
    /usr/local/bin/migrate -path /app/migrations -database "$CONN_STR" up || {
      echo "Migration failed! Exiting."
      exit 1
    }
    echo "Migrations applied successfully!"
  else
    echo "========================================="
    echo "Skipping migrations: No database connection info found (neither DATABASE_URL nor POSTGRES_HOST is set)."
    echo "========================================="
  fi
else
  echo "========================================="
  echo "Skipping migrations: RUN_MIGRATIONS is set to false."
  echo "========================================="
fi

# Execute the CMD (the Go server binary /app/server)
exec "$@"
