# backend

Server-side application module.

## Development Commands

- `go test ./...` - run backend tests.
- `sqlc generate` - generate type-safe query code from SQL contracts.
- `goose -dir ./migrations postgres "$DATABASE_URL" up` - apply pending database migrations.

## Running Locally

1. Set required environment variables (example values):
   ```bash
   export DATABASE_URL="postgres://activist:activist@localhost:5432/activist_base?sslmode=disable"
   export SESSION_IDLE_TTL="8h"
   export SESSION_ABSOLUTE_TTL="168h"
   export HTTP_ADDR=":8080"
   export PUBLIC_BASE_URL="http://localhost:8080"
   export SESSION_COOKIE_NAME="__Host-session"
   export LOG_LEVEL="INFO"
   export SEED_SUPERUSER_LOGIN="admin"
   export SEED_SUPERUSER_PASSWORD="change-me"
   ```
2. Optional development flags:
   ```bash
   export DEV_AUTH_ASSUME_ADMIN="false"
   export DEV_AUTH_ADMIN_LOGIN="admin"
   ```
3. Start the backend:
   ```bash
   go run ./cmd/server
   ```
4. Expected startup output includes `starting http server` with the configured `addr` (for example `:8080`).
5. Health endpoint: `GET /healthz` (returns `204 No Content`).

## Environment Variables

| Variable | Required | Default | Description |
| --- | --- | --- | --- |
| `DATABASE_URL` | Yes | none | PostgreSQL connection string used by the API server and migration runner. |
| `SESSION_IDLE_TTL` | Yes | none | Session idle timeout as Go duration (for example `8h`). |
| `SESSION_ABSOLUTE_TTL` | Yes | none | Maximum session lifetime as Go duration (for example `168h`). |
| `HTTP_ADDR` | Yes | none | HTTP listen address for `http.Server` (for example `:8080`). |
| `PUBLIC_BASE_URL` | Yes | none | purpose unclear - verify before deploying (loaded in config, not used by current server runtime). |
| `SESSION_COOKIE_NAME` | Yes | none | Name of the authentication session cookie. |
| `LOG_LEVEL` | Yes | none | Structured log level parsed by `log/slog` (invalid values fall back to `INFO`). |
| `DEV_AUTH_ASSUME_ADMIN` | No | `false` | Enables development-only admin assumption in auth middleware. |
| `DEV_AUTH_ADMIN_LOGIN` | No | `admin` | Login name used when `DEV_AUTH_ASSUME_ADMIN=true`. |
| `SEED_SUPERUSER_LOGIN` | Yes | none | purpose unclear - verify before deploying (loaded in config, not used by current server runtime). |
| `SEED_SUPERUSER_PASSWORD` | Yes | none | purpose unclear - verify before deploying (loaded in config, not used by current server runtime). |

## Docker Build

```bash
docker build -t backend:latest .
docker build -f Dockerfile.migrate -t backend-migrate:latest .
```

### Production image (tag used by deploy compose)

```bash
docker build -t activist-backend:prod .
docker save activist-backend:prod | gzip > activist-backend-prod.tar.gz
```

Transfer to server and load:

```bash
scp -P 64971 activist-backend-prod.tar.gz deploy@77.221.139.39:/opt/activist/
ssh -p 64971 deploy@77.221.139.39 'docker load < /opt/activist/activist-backend-prod.tar.gz'
```

## Database Migrations

Migration format/tooling is Goose (`-- +goose` directives in `migrations/*.sql`, postgres dialect).

Without Docker:

```bash
go install github.com/pressly/goose/v3/cmd/goose@v3.26.0
goose -dir ./migrations postgres "$DATABASE_URL" up
```

With Docker:

```bash
docker run --rm -e DATABASE_URL="$DATABASE_URL" backend-migrate:latest
```

## API Endpoints

| Method | Path | Auth Required | Description |
| --- | --- | --- | --- |
| `GET` | `/healthz` | No | Liveness/health probe endpoint. |
| `POST` | `/api/v1/auth/register` | No | Register a new user and create a session. |
| `POST` | `/api/v1/auth/login` | No | Authenticate user and create a session. |
| `GET` | `/api/v1/auth/session` | Yes | Inspect current session and actor permissions. |
| `POST` | `/api/v1/auth/logout` | Yes | Revoke current session. |
| `POST` | `/api/v1/auth/logout-all` | Yes | Revoke all sessions for current user. |
| `GET` | `/api/v1/users/{user_id}` | Yes | Read user profile by ID. |
| `PATCH` | `/api/v1/users/{user_id}` | Yes | Update user profile fields. |
| `GET` | `/api/v1/users/{user_id}/memberships` | Yes | List memberships for a user. |
| `GET` | `/api/v1/divisions` | Yes | List root or child divisions (`parent_id` query). |
| `POST` | `/api/v1/divisions` | Yes | Create a division. |
| `PATCH` | `/api/v1/divisions/{division_id}` | Yes | Update division metadata/parent. |
| `POST` | `/api/v1/divisions/{division_id}/archive` | Yes | Archive a division (cascade behavior). |
| `GET` | `/api/v1/divisions/tree` | Yes | Fetch division tree (`depth`, `include_archived` queries). |
| `POST` | `/api/v1/roles` | Yes | Create a role with permission set. |
| `GET` | `/api/v1/roles` | Yes | List roles. |
| `PATCH` | `/api/v1/roles/{roleID}` | Yes | Edit role name/permissions. |
| `DELETE` | `/api/v1/roles/{roleID}` | Yes | Delete role. |
| `POST` | `/api/v1/divisions/{division_id}/positions` | Yes | Create a position in a division. |
| `GET` | `/api/v1/divisions/{division_id}/positions` | Yes | List positions in a division. |
| `GET` | `/api/v1/positions/{position_id}/members` | Yes | List users assigned to a position. |
| `POST` | `/api/v1/positions/{position_id}/archive` | Yes | Archive a position. |
| `POST` | `/api/v1/memberships` | Yes | Assign user to position. |
| `DELETE` | `/api/v1/positions/{position_id}/members/{user_id}` | Yes | Remove user from position. |
| `GET` | `/api/v1/eventlog` | Yes | List audit/event log entries. |
| `GET` | `/api/v1/search/users` | Yes | Search users with query filters and paging. |
