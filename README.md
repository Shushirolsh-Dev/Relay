# Relay

A self-hostable product waitlist with a polished launch page, persistent signups, referral attribution and a private admin console. No paid SaaS account required.

![Go](https://img.shields.io/badge/Go-1.23+-00ADD8?logo=go) ![SQLite](https://img.shields.io/badge/Database-SQLite-003B57?logo=sqlite) ![License](https://img.shields.io/badge/License-MIT-green)

## Features
- Responsive launch page and accessible signup form.
- Go JSON API with email validation, normalization and duplicate protection.
- SQLite persistence; each signup receives a shareable referral code.
- Referral attribution from `?ref=CODE` links.
- Private admin dashboard with totals, recent signups and CSV export.
- Security headers, bounded request bodies and parameterized SQL.
- Docker and Render deployment configurations.

## Run locally
Requires Go 1.23+.

```bash
git clone https://github.com/Shushirolsh-Dev/Relay.git
cd Relay
export ADMIN_TOKEN="$(openssl rand -hex 32)"
go run ./cmd/relay
```
Visit http://localhost:8080 and open http://localhost:8080/admin for the dashboard.

## Configuration
| Variable | Default | Purpose |
|---|---|---|
| `PORT` | `8080` | HTTP port |
| `DATABASE_PATH` | `./data/relay.db` | SQLite file path |
| `ADMIN_TOKEN` | unset | Required bearer token for admin APIs |
| `SITE_NAME` | `Relay` | Brand name |
| `SITE_TAGLINE` | — | Optional site tagline |

Admin endpoints are disabled until `ADMIN_TOKEN` is configured. Use a long random secret and never commit it.

## API
### `POST /api/waitlist`
Body: `{"email":"hello@example.com","referral_code":"RLY-A1B2C3"}`

Returns `201` with a generated referral code and list position. Duplicate emails return `409`; invalid input returns `400`.

### `GET /api/health`
Returns service and database health.

### Admin (Bearer token required)
- `GET /api/admin/stats` — total and daily signups, latest registrations and top referral sources.
- `GET /api/admin/export.csv` — downloadable CSV.

## Deploy
**Render:** create a Blueprint from the included `render.yaml`. It provisions a Go web service, generated admin token and persistent disk. Keep the disk mounted at `/var/data`.

**Docker:**
```bash
docker build -t relay .
docker run --rm -p 8080:8080 -e ADMIN_TOKEN="$(openssl rand -hex 32)" -v relay-data:/var/data relay
```

## Tests
```bash
go test ./...
```

## Structure
`cmd/relay` entrypoint · `internal/store` SQLite persistence · `internal/httpapi` HTTP API · `web` launch page and admin UI.

## Security & scope
Relay does not send marketing emails or expose public signup data. Admin routes require a bearer token; use HTTPS in production. CSV export protects cells from spreadsheet formula injection. Add a consent-aware email provider separately if you need outbound email.

## License
MIT. See [LICENSE](LICENSE).
