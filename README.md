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

Copyright (c) 2026 Euler-B

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
