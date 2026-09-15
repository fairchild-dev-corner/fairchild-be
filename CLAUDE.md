# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

This is a Go module (`fairchild_be`, Go 1.26). There is no test suite in the repo yet.

```bash
# Build (dev/stage only changes an embedded BuildEnv string, not the binary target)
make build-dev
make build-stage
# equivalent to: go build -ldflags "-X 'main.BuildEnv=dev'" -o bin/fairchild_be cmd/main.go

# Run (builds first, then execs bin/fairchild_be)
make run-dev
make run-staging

# DB migrations (golang-migrate, MySQL driver) - reads cmd/migrate/migrations
make migrate-up-dev
make migrate-down-dev
make migrate-up-stage
make migrate-down-stage

# Compile-check / vet everything without producing a binary
go build ./...
go vet ./...
```

Config is loaded from `.env` via `godotenv` (see `config/env.go`) plus a `BuildEnv` ("dev" or "stage") that selects between the `*_LOCAL_*` and `*_STAG_*` MySQL credential sets in `config/mysql_config.go`. There's no `go test` target defined anywhere yet - if you add tests, run them with the standard `go test ./...` / `go test ./path/to/pkg -run TestName`.

## Architecture

**Entry point chain**: `cmd/main.go` loads all config structs (`config/*.go`) and the MySQL pool, then calls `config.ServerConfig(...)` (`config/server_config.go`), which constructs the gin engine and calls `core.NewAPIServer(...).Run()` (`cmd/api/api.go`, package `core`). `Run()` is where the actual dependency wiring happens for each feature area:

```
repo := auth_repo.NewAuthRepository(db)          // internal/repositories/auth
service := auth_service.NewAuthService(repo, ...) // internal/services/auth
handler := auth_handler.NewHandler(service)       // internal/handlers/auth
handler.RegisterRoutes(routesGroup)
```

Every feature area (currently just `auth`) follows this same repo → service → handler layering, wired by hand in `cmd/api/api.go` (no DI framework). When adding a new feature area, follow the same three-layer split and wire it into `APIServer.Run()`.

All routes are mounted under `/api/v1` (`internal/routes`, package `router`), which also applies global CORS middleware and an error-collecting middleware (`loggers.ErrorHandler()`). Auth routes additionally get a per-IP token-bucket rate limiter (`internal/middlewares/ip_rate_limiter.go`) applied at the route-group level - see `auth_routes.go` / `sso_handler.go` for the burst/refill constants.

### Auth module structure (`internal/{handlers,services,repositories,models}/auth`)

- **Two-factor login**: `POST /auth/login` (member_id + password) never returns tokens directly - on success it creates an OTP challenge row (`otp_verifications` table), sends the code via SMS (Movider) and email (SMTP, best-effort secondary channel), and returns a `reference_id`. `POST /auth/login/verify-otp` redeems that reference_id + code and only then issues JWT access/refresh tokens. The same OTP-challenge machinery (generate → store hash → send → verify → consume) is reused for other flows (e.g. forgot-password) by varying the `purpose` column/field on `OTPVerification`.
- **Social login vs SSO are two distinct handlers**: `AuthHandler.handleSocialLogin` (`POST /auth/social/login`) re-verifies a token the client already obtained from a provider's native SDK (see `*_verifier.go` files implementing `SocialVerifier`). `SSOHandler` (`sso_handler.go`) instead drives a server-initiated OAuth2 redirect via `goth`/`gothic` (`GET /auth/sso/:provider` and `.../callback`), completing with a redirect back to the frontend carrying the access token in the URL fragment (never a query string, so it's not logged/sent to the server) - the refresh token instead goes out as an HttpOnly `Set-Cookie` on that same redirect response, never JS-readable. Both handlers converge on `FindLinkedMemberFromSocialProfile` and `AuthService.issueTokens`. `FindLinkedMemberFromSocialProfile` only *links* a verified social identity to an existing member found by email - it deliberately never creates a new `users` row, so an OAuth login from an email with no matching member fails with `cc.ErrEmailNotAssociated` (surfaced by the SSO redirect as `#error=not_a_member`) rather than silently provisioning a new active account with no `member_id`, no OTP/mobile verification, and no staff review. Both social-login paths also reject `UserStatusSuspended` members the same way password login does.
- **Sessions**: refresh tokens are opaque random values (`GenerateOpaqueToken`); only their SHA-256 hash is persisted (`auth_sessions.refresh_token_hash`). Refresh is rotate-on-use: `RefreshTokenService` revokes the presented token and issues a brand new pair.
- **Login lockout**: `CountRecentFailedLogins` checks both the member_id and the source IP against a rolling window computed in SQL (not in Go), and is re-checked at both the password step and the OTP-verify step so a mix of bad passwords and bad OTP guesses across either dimension still locks the account/IP out.
- **Error handling convention**: sentinel errors live in `internal/constants/error_constants.go` (`cc.Err...`), checked via `errors.Is` in handlers to pick an HTTP status; unwrapped/dynamic errors compare `err.Error()` against the matching message constant. Security-sensitive flows (OTP verification, login) deliberately collapse multiple failure reasons (wrong code, expired, exhausted attempts, unknown reference) into one generic sentinel so a client response can't be used to narrow down which case occurred.
- **Response shape**: successful responses go through `loggers.StatusOK`/`loggers.SuccessResponse`, wrapping the body in `common.SuccessResponse` with a `Status` message from `internal/constants/message.go` (`cc.SUCCESS_*`). Errors go through `loggers.GetCommonError(ctx, message, httpStatus)`, wrapping in `common.ErrorResponse`.
- **Request validation**: handlers call `utils.ValidatePayload(ctx, &req)` once, which binds JSON and runs `go-playground/validator` struct tags in one step; a `false` return means the response was already written and the handler should return immediately.
- **Migrations**: `cmd/migrate/migrations/*.up.sql` / `*.down.sql`, applied via `golang-migrate`. Table renames/column additions happen as new numbered migration pairs rather than editing old ones (e.g. `otp_challenges` was later renamed to `otp_verifications`, then got a `status`/`send_error` pair added in a follow-up migration) - follow that pattern for schema changes.
