# core

Library for building web applications and microservices.

built this to stop wasting my time writing over and over again the setup in every project

this repository collects small utilities and middleware that are commonly needed when building web services: configuration loading, db/cache helpers.

## layout

- `cctx/` — typed context keys and small context helpers used across middleware and handlers.
- `middleware/` — HTTP middlewares (targetted to use with chi)
  - `middleware.go` — request/response helpers, JSON writer, common middlewares (logger, recoverer, RealIP extraction, internal-only guard, role-based allow). Includes a `Wrap` helper that turns handlers returning errors into standard HTTP handlers.
  - `ratelimiter.go` — per-IP token-bucket rate limiter using `golang.org/x/time/rate` with automatic cleanup.
- `utils/` — small helpers:
  - `errs` — `ApiError` type used through `Wrap` for shaping HTTP error responses.
  - `jsonutil` — JSON helpers and request validation integration (`go-playground/validator`).
  - `utils` — common helpers

## Quick start

1. Wire the provided middleware into a chi router and use `middleware.Wrap` for handlers that return errors. Example:

```go
r := chi.NewRouter()

// common middlewares
r.Use(middleware.Logger)
r.Use(middleware.Recoverer)

// handlers that return error can be used with Wrap
r.Get("/", middleware.Wrap(func(w http.ResponseWriter, req *http.Request) error {
    w.Write([]byte("ok"))
    return nil
}))
```

2. Return `*errs.ApiError` from wrapped handlers when you need to control HTTP response codes.

## Design notes

- `jsonutil.Parse` uses `go-playground/validator` for request payload validation. Define struct tags to validate input.

## Testing and quality

This repo is intentionally small and minimal. It relies on battle-tested upstream libraries for transport, DB drivers and validation. Add unit tests in your service that uses these helpers; the helpers themselves are thin wrappers and straightforward to test with small integration tests (e.g., local Redis or testcontainers).
