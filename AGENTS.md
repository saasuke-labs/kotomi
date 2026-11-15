# Kotomi: agent notes

Kotomi is a self-hosted Go service that adds comments and reactions to static sites. It has three parts: a REST API, a server-rendered admin panel for site owners (moderation, configuration, analytics), and a dependency-free JavaScript widget that a static page embeds. Module `github.com/saasuke-labs/kotomi`; `VERSION` is `0.1.0-beta.1` (beta, not production-hardened).

## Ecosystem (Saasuke Labs)

- **Gengo** (`saasuke-labs/gengo`, static site generator) is the main intended host. Integration is loose: a Gengo page template includes `kotomi.css` and `kotomi.js` from a Kotomi server and calls `new Kotomi({baseUrl, siteId, pageId, container}).render()`. There is no code dependency in either direction. Guides: `docs/GENGO_QUICK_START.md`, `docs/GENGO_INTEGRATION_GUIDE.md`, example `frontend/examples/gengo-example.html`.
- **Nagare** (`saasuke-labs/nagare`, text-to-SVG diagrams) is unrelated to this codebase.
- **saasuke-labs.com** is built from `saasuke-labs/site-source` (published to `saasuke-labs/saasuke-labs.github.io`). `docs/index.html` is Kotomi's product page in that site's design: it loads CSS, fonts and JS from `saasuke-labs.com/static/` and links to `/gengo/` and `/nagare/`. It is hand-written HTML; keep it in sync when user-facing behaviour changes.
- This repo has no release workflow and no step that notifies `site-source` (nagare's release workflow dispatches to it). Nothing here triggers a site rebuild when Kotomi is released.

## Commands

Go toolchain: `go.mod` requires `go 1.25.0`. With an older Go and the default `GOTOOLCHAIN=auto`, the first `go` command downloads 1.25.0 (needs network). CGO is required (`github.com/mattn/go-sqlite3`), so `gcc` must be present.

```bash
go build -o kotomi ./cmd/main.go          # binary name is gitignored
go run ./cmd/main.go                      # must run from the repo root (see gotchas)
PORT=8765 DB_PATH=/tmp/k.db go run ./cmd/main.go

go test ./...                             # all unit + integration tests; e2e self-skips
go test ./pkg/middleware -run TestRateLimit -count=1     # single test
go test ./pkg/comments/ -run Integration -v              # what CI calls "integration"
RUN_E2E_TESTS=true go test ./tests/e2e/... -v -timeout=10m   # starts `go run cmd/main.go` on :8888 with a temp DB
go vet ./...

bash frontend/build.sh                    # widget: frontend/src -> frontend/dist
swag init -g cmd/main.go -o docs          # regenerate Swagger (needs github.com/swaggo/swag/cmd/swag; not installed by default)
docker build -t kotomi .                  # multi-stage, static CGO binary on alpine
```

There is no Makefile or justfile. `.env` is not loaded by the program; export variables yourself (`set -a; . ./.env; set +a`).

Environment variables read by the code (complete list):

| Variable | Default | Purpose |
| --- | --- | --- |
| `PORT` | `8080` | Listen port |
| `DB_PROVIDER` | `sqlite` | `sqlite` or `firestore` |
| `DB_PATH` | `./kotomi.db` | SQLite file |
| `FIRESTORE_PROJECT_ID` / `GCP_PROJECT` | none | Firestore project (credentials via ADC or `GOOGLE_APPLICATION_CREDENTIALS`) |
| `AUTH0_DOMAIN`, `AUTH0_CLIENT_ID`, `AUTH0_CLIENT_SECRET` | none | All three enable the admin panel, `/login`, and `/swagger/` |
| `AUTH0_CALLBACK_URL` | `http://localhost:8080/callback` | Admin OAuth callback (`.env.example` wrongly shows `/auth/callback`) |
| `SESSION_SECRET` | dev fallback | Admin cookie store key; `ENV=production` sets `Secure` cookies |
| `OPENAI_API_KEY` | none | Enables `OpenAIModerator`; otherwise `MockModerator` |
| `CORS_ALLOWED_ORIGINS`, `CORS_ALLOWED_METHODS`, `CORS_ALLOWED_HEADERS`, `CORS_ALLOW_CREDENTIALS` | `*`, `GET,POST,PUT,DELETE,OPTIONS`, `Content-Type,Authorization`, `false` | API CORS |
| `RATE_LIMIT_GET`, `RATE_LIMIT_POST` | `100`, `5` | Requests per minute per client IP |

With no variables set the server starts on SQLite with the admin panel disabled; `GET /healthz` returns `{"message":"OK"}`.

## Architecture

Request flow: `cmd/main.go` (logger, `db.NewStore`, Auth0 config, templates, moderator, notification queue, graceful shutdown) builds a `server.Config` → `cmd/server/server.go` (`Server`, `Handler()`) → `cmd/server/routes.go` (`RegisterRoutes`, the single route table, gorilla/mux) → `cmd/server/handlers/` (public API: `comments.go`, `reactions.go`, admin login/dashboard in `auth.go`) and `pkg/admin/` (admin panel handlers).

Routes:

- `GET /healthz`, `GET /` (serves `static/index.html`), `/static/*` (file server over `./static`).
- `/api/v1/...` with CORS and rate limiting. Public reads: `site/{siteId}/page/{pageId}/comments`, `site/{siteId}/allowed-reactions`, `site/{siteId}/comments/{commentId}/reactions[/counts]`, `site/{siteId}/pages/{pageId}/reactions[/counts]`. JWT-protected writes: POST comment, PUT/DELETE `site/{siteId}/comments/{commentId}` (owner only), POST reactions, DELETE `site/{siteId}/reactions/{reactionId}`. Note the path uses `page/` for comments and `pages/` for reactions.
- `/api/...` (unversioned) mirrors the same handlers behind `handlers.DeprecationMiddleware`. Add new endpoints to `/api/v1` only.
- `/api/v1/auth/{login,callback,logout,user,config}`: end-user login for "kotomi" auth mode (`pkg/auth/handlers.go`).
- `/login`, `/callback`, `/logout`, `/admin/...`, `/swagger/`: registered only when Auth0 is configured. Admin routes sit behind `auth.RequireAuth` (cookie session).

Packages:

| Package | Owns |
| --- | --- |
| `pkg/db` | `Store` interface (comment CRUD + `GetDB() *sql.DB`), `ConfigFromEnv`, `NewStore` factory, `SQLiteAdapter`, `FirestoreStore` |
| `pkg/comments` | `Comment` struct, `SQLiteStore` (opens the DB, sets WAL pragmas, **creates the whole schema** for every table), legacy in-memory `SitePagesIndex` |
| `pkg/models` | SQL stores for everything except comments: `SiteStore`, `PageStore`, `AdminUserStore`, `UserStore`, `AllowedReactionStore`, `ReactionStore`, `SiteAuthConfigStore`; `KotomiUser` (JWT claims) |
| `pkg/auth` | Admin Auth0 OAuth + gorilla cookie sessions (`auth0.go`, `middleware.go`); `JWTValidator` (hmac/rsa/ecdsa/jwks); kotomi-mode users, sessions and token issuing (`kotomi_auth.go`, `handlers.go`) |
| `pkg/middleware` | `RequestIDMiddleware`, `LoggingMiddleware`, CORS (`rs/cors`), token-bucket `RateLimiter`, `JWTAuthMiddleware` / `GetUserFromContext` |
| `pkg/admin` | Admin handlers: sites, pages, comments (approve/reject/bulk), reactions, moderation, notifications, per-site auth config, users, export/import, analytics |
| `pkg/moderation` | `Moderator` interface, `OpenAIModerator`, `MockModerator`, `DetermineStatus`, per-site `ConfigStore` |
| `pkg/notifications` | SQL-backed email queue (`Queue.Start` polls every 30s), SMTP and SendGrid providers, settings store, email templates |
| `pkg/analytics`, `pkg/export`, `pkg/import` | Admin analytics queries; JSON/CSV export and import |
| `pkg/errors`, `pkg/logging` | `APIError` JSON envelope (`WriteErrorWithRequestID`); slog `ContextHandler` that adds request/site/page/comment IDs from context |

Storage: only comments go through `db.Store`. Everything else (sites, pages, reactions, auth config, users, moderation config, notifications, analytics, the whole admin panel, and `JWTAuthMiddleware`) uses the raw `*sql.DB` from `Store.GetDB()`, which is `nil` for Firestore. `cmd/main.go` disables moderation config and notifications when it is nil, but the other code paths do not guard against it. Treat Firestore as a comments-only backend that is not wired end to end; SQLite is the working configuration and what all tests use.

Data model (SQLite, defined in `pkg/comments/sqlite.go`): `admin_users` → `sites` (owner) → `pages` (`UNIQUE(site_id, path)`) → `comments` (`site_id`, `page_id`, `parent_id` for replies, `status` in `pending|approved|rejected`) → `reactions` (exactly one of `page_id` / `comment_id`, plus `allowed_reaction_id`, `user_id`). `allowed_reactions` are per site with `reaction_type` `page|comment|both`. `users` are per site, keyed `(site_id, id)`, upserted from JWT claims on every authenticated request. Also `site_auth_configs`, `kotomi_auth_users`, `kotomi_auth_sessions`, `moderation_config`, `notification_settings`, `notification_queue`, `notification_log`. `SQLiteStore.AddPageComment` auto-creates a `system` admin user, the site and the page when they do not exist.

Auth, two separate systems:

- **Admin** (site owners): Auth0 OAuth, cookie session, `auth.RequireAuth`. First login creates an `admin_users` row.
- **End users** (commenters): per-site `site_auth_configs` row, managed at `/admin/sites/{siteId}/auth`. `auth_mode=external`: the host site issues JWTs and Kotomi validates them by `jwt_validation_type` (`hmac`, `rsa`, `ecdsa`, `jwks`), issuer and audience. `auth_mode=kotomi`: Kotomi logs users in through Auth0 at `/api/v1/auth/*` and issues its own HMAC token (also read from the `kotomi_auth_token` cookie). Tokens must carry `sub` and a `kotomi_user` object with at least `name` (`extractUserFromClaims`). A site without an auth config rejects all writes with 401. `scripts/generate_jwt.{js,py}` produce test tokens.

Moderation: `PostComments` calls `Moderator.AnalyzeComment` only when the site's `moderation_config.enabled` is set, then `DetermineStatus` maps confidence to approved/pending/rejected using the site thresholds. Default status is `pending`.

Admin UI: Go `html/template` files in `templates/` with HTMX (loaded from unpkg) and `static/css/admin.css`, `static/js/charts.js`. Several handlers return an HTML partial when `HX-Request: true` and JSON otherwise.

Widget: sources in `frontend/src/{api.js,ui.js,styles.css}`; `frontend/build.sh` concatenates them into `frontend/dist/kotomi.js` (+ `.min.js`, `.css`, `.min.css`). `dist/` is committed and the build is reproducible (rebuilding yields no diff). The server serves the widget from `static/kotomi.*`, which are byte-identical copies of `frontend/dist/*`.

## Conventions and gotchas

- **Working directory matters.** `templates/` and `static/` are read by relative path at runtime, so the binary must run from the repo root (the Docker image copies both next to the binary).
- **Admin template loading looks broken.** `cmd/main.go` calls `template.ParseGlob("templates/**/*.html")`; Go globs have no recursive `**`, so only `templates/admin/dashboard.html` is parsed (as `dashboard.html`) and the per-file fallback never runs. Handlers then execute names such as `admin/dashboard.html`, `login.html`, `sites/list.html`, and template `define` names are inconsistent (`reactions/list.html` defined, `admin/reactions/list.html` requested). Verified: `GET /admin` without Auth0 returns 500 "Template error". No test covers admin HTML rendering. Check this before trusting any admin page, and fix loading and naming together.
- Kotomi-mode JWTs are signed with a hard-coded secret, `"kotomi-internal-jwt-secret-change-in-production-" + siteID`, duplicated in `pkg/auth/handlers.go` (`GetJWTSecret`) and `pkg/auth/jwt_validator.go` (`validateKotomiToken`). Keep both in sync until it moves to configuration.
- `GET .../comments` returns comments of every status; neither store filters on `approved`.
- Widget build: `frontend/src/kotomi.js` is read by `build.sh` but never used. The `Kotomi` class that ships is the heredoc inside `build.sh`, so change it there (and mirror `src/kotomi.js`). After building, copy `frontend/dist/kotomi*` to `static/` by hand; nothing automates it. The minified JS is a plain copy.
- Schema changes go in the single `CREATE TABLE IF NOT EXISTS` block in `pkg/comments/sqlite.go`. There is no migration system, so existing databases do not pick up column changes.
- Generated, do not hand-edit: `docs/docs.go`, `docs/swagger.json`, `docs/swagger.yaml` (from the `// @...` annotations on handlers and `cmd/main.go`), `frontend/dist/*`, `static/kotomi*`. `docs/` is both a Go package (imported by `routes.go`) and the home of prose docs and `docs/index.html`.
- Dead files: `cmd/main.go.backup`, `cmd/main.go.old` (pre-refactor monolith), `cmd/main_test.go.disabled`, `cmd/reactions_test.go.disabled`. `cmd/` currently has no tests. `InMemoryStoreAdapter` in `cmd/main.go` and `comments.SitePagesIndex` are unused leftovers. `TEST_MODE` is set by CI and e2e but read nowhere.
- The Dockerfile passes `-X main.version/commit/buildTime`, but `cmd/main.go` declares none of them, so the flags do nothing. The running binary does not know its version.
- 57 Go files are not `gofmt`-clean. Format only the code you touch; a repo-wide `gofmt -w` would bury the real change.
- Rate limiting keys on `X-Forwarded-For` (first entry), then `X-Real-IP`, then `RemoteAddr`. E2E raises both limits to 1000.
- Docs mention a `ghcr.io/saasuke-labs/kotomi:latest` image; no workflow here publishes to GHCR (CI pushes to GCP Artifact Registry).
- ADRs in `docs/adr/` are all marked "Proposed", but 002 (the `cmd/server` + handlers split, Go 1.25) and 003 (SQLite WAL, pool settings) are implemented. Follow them: handlers are methods on `ServerHandlers`, dependencies are injected through `server.Config`, and the `db.Store` and `pkg/models` stores take `context.Context` (the notification and analytics stores do not).
- API errors use `pkg/errors` (`apierrors.WriteErrorWithRequestID`); log with `s.Logger.*Context(ctx, ...)` after enriching `ctx` via `logging.WithSiteID` etc.

## CI, release, deploy

- One workflow, `.github/workflows/deploy_kotomi.yaml`, on push and PR to `main`: `unit-tests` (`go test ./pkg/... ./cmd/... -race`), `integration-tests`, `e2e-tests`, all on Go 1.25.x.
- On push to `main`, after tests pass, `build-and-deploy` builds the Docker image, pushes it to GCP Artifact Registry tagged with the contents of `VERSION`, and deploys Cloud Run service `kotomi-prod` (secrets `GCP_SA_KEY`, `GCP_PROJECT_ID`, `GCP_REGION`). **Every merge to `main` deploys to production**, overwriting the image tag unless `VERSION` was bumped. The deploy step sets no application env vars; they live on the Cloud Run service.
- Cutting a version is manual: edit `VERSION`, add a section to `CHANGELOG.md` (Keep a Changelog), tag `vX.Y.Z`, create the GitHub release. See `docs/RELEASE_PROCESS.md`.

## Which docs to trust

| Topic | Source |
| --- | --- |
| Routes, env vars, schema | The code: `cmd/server/routes.go`, `pkg/middleware`, `pkg/comments/sqlite.go` |
| API reference | `docs/swagger.yaml` (regenerate after changing annotations); `README.md` has long curl examples |
| End-user JWT format | `docs/AUTHENTICATION_API.md`, `docs/KOTOMI_AUTH.md`, ADR 001 |
| Deploy and release | `docs/GCP_DEPLOYMENT_SETUP.md`, `docs/RELEASE_PROCESS.md`, `docs/FIRESTORE_SETUP.md`, `docs/DATABASE_BACKUP_RESTORE.md` |
| Admin panel usage | `docs/ADMIN_PANEL_GUIDE.md` |
| Widget | `frontend/README.md`, `frontend/examples/` |
| Manual requests | `.http/requests.http`, `scripts/smoke_test.sh`, `scripts/smoke_test_authenticated.sh` |

Historical snapshots, not current truth: `Status.md` (still says v0.0.1), `AUTHENTICATION_STATUS_REPORT.md`, `DATABASE_ABSTRACTION_REFACTORING.md`, `DEPENDENCY_DIAGRAM.md`, `FIRESTORE_IMPLEMENTATION_SUMMARY.md` (claims the whole codebase is backend-agnostic; it is not), `SQLITE_OPTIMIZATION_SUMMARY.md`, `TASK_SUMMARY.md`, `TASK_COMPLETION_SUMMARY.md`, `TESTING_REPORT.md`, `ISSUES_SUMMARY.md`, `CREATING_ISSUES.md`, `docs/PHASE*_*.md`, `docs/BETA_*.md`, `frontend/IMPLEMENTATION_SUMMARY.md`, `internal_docs/` (early requirements and data model), `scripts/create_issues.sh`.
