# Telurify Ingestion

Periodic Go ingestion job for the USGS earthquake GeoJSON feed. It ports the
existing Rails rake task and writes directly to the `sismos` table used by
`Telurify-API`.

## Local execution

Go does not run database migrations. Configure `DATABASE_URL` in `.env` or your shell environment to point to a PostgreSQL database with the `sismos` table already created. Then run:

```bash
source .env
make run
```

The application emits structured, colorized logs to stderr when running in a
terminal. Colors are disabled automatically when output is redirected. The
`DATABASE_URL` value can also be supplied directly in the shell environment.

The job intentionally deduplicates by `title` and validates magnitudes in
`-1.0..10.0`. Database lookup and insert errors are reported separately from
actual duplicates.

### Make targets

```bash
make help  # Show available commands
make test   # Run tests
make vet    # Run static analysis
make check  # Format, test, and vet
make build  # Build bin/ingest
```

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
