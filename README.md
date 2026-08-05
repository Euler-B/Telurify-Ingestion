# Telurify Ingestion

Periodic Go ingestion job for the USGS earthquake GeoJSON feed. It ports the
existing Rails rake task and writes directly to the `sismos` table used by
`Telurify-API`.

## Local execution

Go does not run database migrations. Configure `DATABASE_URL` in `.env` or your shell environment to point to a PostgreSQL database with the `sismos` table already created. Then run:

```bash
go mod tidy
DATABASE_URL="postgresql://telurify:telurify@localhost:5432/backend_development?sslmode=disable" go run ./cmd/ingest
```

The job intentionally deduplicates by `title`, validates magnitudes in
`-1.0..10.0`, and classifies other save errors as duplicates to preserve the
Rails task's behavior.

### Database permissions

For production, use a dedicated PostgreSQL role for this job instead of the
Rails owner role. Create it with an administrator and store its connection
string as the GitHub Actions secret `DATABASE_URL`:

Do not grant this role `UPDATE`, `DELETE`, `CREATE`, or migration privileges.
Use `sslmode=require` (or stronger verification) in the production connection
string.

## GitHub Actions

The workflow runs every six hours and can also be started with
`workflow_dispatch`. Configure the Neon connection string as the repository
secret `DATABASE_URL` before running it.

## License

This project is licensed under the [MIT License](LICENSE).
