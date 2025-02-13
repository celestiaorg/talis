# Database Migrations

This project uses `golang-migrate` for database migrations.

## Creating New Migrations

```bash
migrate create -ext sql -dir migrations -seq name_of_migration
```

## Running Migrations

```bash
# Apply all pending migrations
make migrate

# Rollback all migrations
make migrate-down

# Force a specific version
make migrate-force version=1

# Check current version
make migrate-version
```

## Migration Files

- `000001_create_jobs_table.up.sql`: Initial jobs table
- `000002_add_webhook_details.up.sql`: Added webhook functionality 