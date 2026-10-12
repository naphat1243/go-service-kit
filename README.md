# go-service-kit

**A production-grade Go REST service you can ship on day one — clean architecture, real tests, Docker-ready, and built to be extended by AI coding agents.**

Most Go boilerplates give you a folder structure. This one gives you the first two weeks of work: typed error handling wired to HTTP, a full CRUD slice through every layer, graceful shutdown, migrations, rate limiting, request logging, CI — all working out of the box.

```bash
git clone https://github.com/naphat1243/go-service-kit && cd go-service-kit
make dev        # starts postgres, runs migrations, boots the API
```

```bash
curl -X POST localhost:8080/api/v1/users \
  -H 'Content-Type: application/json' \
  -d '{"email":"jo@example.com","name":"Jo"}'
```

That's it — a real service talking to a real database in under 5 minutes.

## What's inside

| | |
|---|---|
| **HTTP** | Echo v4 · request ID · panic recovery · rate limiting · structured request logs (slog) |
| **Errors** | Typed `apperror` package — services return domain errors, one place maps them to HTTP. No hand-rolled status codes in handlers, no leaked internals |
| **Data** | PostgreSQL via pgx v5 · SQL migrations (golang-migrate) · repository interfaces owned by the service layer |
| **Example slice** | `users` CRUD through every layer — handler → service → repository — the pattern you copy for every new resource |
| **Tests** | Table-driven unit tests with hand-written fakes (no mock framework) — `make test` passes with no database |
| **Ops** | Graceful shutdown · `/healthz` + `/readyz` (readiness pings the DB) · multi-stage Dockerfile (non-root, ~15MB) · docker-compose · GitHub Actions CI |

## Layout

```
cmd/api/              entry point — config, DI wiring, graceful shutdown
internal/
  config/             env config with sane local defaults
  http/               server, middleware, error handler wiring
  handler/            HTTP layer only: parse request, shape response
  service/            business logic — depends on interfaces, fully unit-tested
  repository/         pgx implementations + SQL migrations
  pkg/apperror/       typed errors → HTTP mapping
```

**The rule that keeps this maintainable:** dependencies point inward. Handlers know services; services know repository *interfaces*; nothing knows Echo except `internal/http` and handlers. Swap Postgres for anything by reimplementing one interface.

## Commands

```bash
make dev           # postgres up + migrate + run
make test          # unit tests (no DB needed)
make lint          # go vet + gofmt check
make migrate-up    # run migrations against DATABASE_URL
make build         # static binary at bin/api
```

## Adding a resource (the 10-minute loop)

1. Migration: `internal/repository/migrations/000X_create_things.up.sql`
2. Service: define the model + repository interface + logic in `internal/service/thing.go`, with tests against a fake
3. Repository: implement the interface in `internal/repository/thing.go`
4. Handler: `internal/handler/thing.go`, register in `cmd/api/main.go`

Every layer has a worked example (`users`) to copy from.

## Pro version — coming soon

The free core is MIT and stays free. **go-service-kit Pro** adds the production features that take real weeks:

- 🔐 JWT auth (access/refresh), password reset, RBAC middleware
- 💳 Stripe subscription billing with webhook handling
- 📈 Prometheus metrics + OpenTelemetry tracing + audit log
- ☸️ Helm chart + staging/prod values + CD pipeline
- 🧪 Integration test suite (testcontainers) covering auth & billing flows
- 🤖 **AI Pack**: a tuned `CLAUDE.md` + agent skills so Claude Code / Cursor extends this codebase *correctly* — conventions enforced, QA checklist included

**⭐ Star this repo to follow the Pro release** — launching on Gumroad soon.

## License

MIT © naphat1243
