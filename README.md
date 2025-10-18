# core

Library for building web applications and microservices.

built this to stop wasting my time writing over and over again the setup in every project

this repository collects small utilities and middleware that are commonly needed when building web services: configuration loading, db/cache helpers.

## layout

- `config/` — environment-driven application configuration helpers (uses .env)
- `database/` — database utilities
- `middleware/` — HTTP middlewares (targetted to use with chi)
  - `middleware.go` — request/response helpers, JSON writer, common middlewares (logger, recoverer, RealIP extraction, internal-only guard, role-based allow). Includes a `Wrap` helper that turns handlers returning errors into standard HTTP handlers.
  - `ratelimiter.go` — per-IP token-bucket rate limiter using `golang.org/x/time/rate` with automatic cleanup.
- `utils/` — small helpers:
  - `errs` — `ApiError` type used through `Wrap` for shaping HTTP error responses.
  - `jsonutil` — JSON helpers and request validation integration (`go-playground/validator`).
  - `utils` — common helpers

## Quick start

1. Create a `.env` (optional) or set environment variables used by config:

   - `ADDR` (default `:8080`)
   - `SECRET_KEY`
   - `DATABASE_URI`
   - `DATABASE_MAX_OPEN_CONNS`, `DATABASE_MAX_IDLE_CONNS`, `DATABASE_MIN_IDLE_CONNS`, `DATABASE_CONN_MAX_LIFETIME`

2. Build components and wire them into your app (example sketch):

```go
cfg := config.New()

pgPool, err := database.NewPostgres(cfg.Database)
if err != nil {
  log.Printf("ehhh: %w", err)
}

cache := database.NewRedisCache(0, "addr")

r := chi.NewRouter()
r.Use(middleware.Logger)
r.Use(middleware.Recoverer)

r.With(middleware.Wrap).Get("/", func(w http.ResponseWriter, req *http.Request) error {
  w.Write([]byte("ok"))
  return nil
})

```

3. Use `middleware.Wrap` to convert handlers that return `error` into `http.HandlerFunc` and return `*errs.ApiError` when appropriate to control response codes.

## Design notes

- `jsonutil.Parse` uses `go-playground/validator` for request payload validation. Define struct tags to validate input.
- Middleware values use typed context keys declared in `middleware/middleware.go` to reduce collisions.

## Testing and quality

This repo is intentionally small and minimal. It relies on battle-tested upstream libraries for transport, DB drivers and validation. Add unit tests in your service that uses these helpers; the helpers themselves are thin wrappers and straightforward to test with small integration tests (e.g., local Redis or testcontainers).
