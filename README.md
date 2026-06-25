# Ticket System (Golang Backend Intern Assignment)

A small backend service where users can register, log in, create tickets, view only
their own tickets, and update the status of their own tickets.

## Tech Stack & Design Choices

- **Language:** Go 1.22 (uses the standard library's enhanced `net/http.ServeMux`
  for method + path-parameter routing, e.g. `GET /tickets/{id}`).
- **Storage:** In-memory (a mutex-guarded map). Simple, meets the assignment's
  scope ("no complex database schema required"), and avoids extra moving parts
  for a two-day project. Data resets on restart — see "Assumptions" below.
- **Auth:** JWT (HS256), implemented with the standard library
  (`crypto/hmac`, `crypto/sha256`, `encoding/json`, `encoding/base64`) rather
  than a third-party JWT package, so the project has **zero external
  dependencies** — no `go.sum`, nothing to fetch from the internet to build.
- **Passwords:** Never stored in plaintext. Stored as `salt:hash` where the hash
  is SHA-256 run for 100,000 rounds over `salt + password` (a simple stdlib-only
  PBKDF-style scheme). Verification uses a constant-time comparison.

No third-party Go modules are used at all, so `go build` works completely
offline with just the Go toolchain.

## Project Structure

```
.
├── main.go        # routing + server startup
├── handlers.go    # HTTP handlers + auth middleware
├── store.go        # thread-safe in-memory data store
├── models.go      # User / Ticket structs + status transition rules
├── jwt.go         # minimal HS256 JWT encode/decode
├── password.go    # salted, iterated SHA-256 password hashing
├── id.go          # random hex ID generator + sentinel errors
├── context.go     # request-context helpers for the authenticated user ID
├── Dockerfile
├── .env.example
└── README.md
```

## API

| Method | Endpoint              | Auth required | Purpose                        |
|--------|------------------------|----------------|---------------------------------|
| GET    | `/health`              | No             | Health check                    |
| POST   | `/auth/register`       | No             | Register a user                 |
| POST   | `/auth/login`          | No             | Log in, returns a JWT            |
| POST   | `/tickets`             | Yes            | Create a ticket                 |
| GET    | `/tickets`             | Yes            | List the logged-in user's tickets |
| GET    | `/tickets/{id}`        | Yes            | Get one of your own tickets      |
| PATCH  | `/tickets/{id}/status` | Yes            | Update status of your own ticket |

Protected endpoints require `Authorization: Bearer <token>`.

### Request/response examples

**Register**
```
POST /auth/register
{"email": "alice@example.com", "password": "password123"}

201
{"id": "...", "email": "alice@example.com"}
```

**Login**
```
POST /auth/login
{"email": "alice@example.com", "password": "password123"}

200
{"token": "<jwt>"}
```

**Create ticket**
```
POST /tickets
Authorization: Bearer <jwt>
{"title": "My printer is broken", "description": "Out of ink"}

201
{"id": "...", "user_id": "...", "title": "...", "description": "...", "status": "open", "created_at": "...", "updated_at": "..."}
```

**Update status**
```
PATCH /tickets/{id}/status
Authorization: Bearer <jwt>
{"status": "in_progress"}

200  (or 400 if the transition is not allowed)
```

### Status flow

```
open -> in_progress -> closed
```

A closed ticket can never move back to `open` or `in_progress`. Any other
transition (e.g. `open` -> `closed` directly) is also rejected with `400`.

### Ownership rules

- A user can only ever see/list/update tickets they created.
- Requesting another user's ticket by ID returns `404` (not `403`), so the
  existence of other users' tickets is never leaked.

## Running Locally (without Docker)

Requires Go 1.22+.

```bash
go build -o ticket-system .
JWT_SECRET=some-long-random-secret ./ticket-system
curl http://localhost:8080/health
```

## Running with Docker

```bash
docker build -t ticket-system .
docker run -p 8080:8080 -e JWT_SECRET=some-long-random-secret ticket-system
curl http://localhost:8080/health
```

Expected response:
```json
{"status": "ok"}
```

## Environment Variables

See `.env.example`.

| Variable     | Required | Default                  | Description                          |
|--------------|----------|---------------------------|---------------------------------------|
| `JWT_SECRET` | No*      | `dev-secret-change-me`    | Secret used to sign/verify JWTs       |
| `PORT`       | No       | `8080`                    | Port the HTTP server listens on       |

\* Not required to run, but **must** be set to a real secret for any
deployment — the app logs a warning if it falls back to the default.

## Deployment

Deployed at: **`<FILL IN YOUR DEPLOYED URL HERE>`**
Public health check: **`<FILL IN>/health`**

> Note: the exact "free-tier" platform you use isn't something I can pick or
> deploy for you from here — I don't have the ability to spin up your own
> cloud account. The two easiest no-cost options for a single Go binary +
> Dockerfile are **Render.com** (free web service, auto-builds from
> Dockerfile) and **Fly.io** (free allowance, `fly launch` detects the
> Dockerfile automatically). Steps for Render:
> 1. Push this repo to GitHub.
> 2. On Render: New → Web Service → connect the repo → it auto-detects the
>    Dockerfile.
> 3. Set the `JWT_SECRET` environment variable in the Render dashboard.
> 4. Render auto-assigns the port via `$PORT`; this app already reads `PORT`
>    from the environment, so no code change is needed.
> 5. Once deployed, confirm `https://<your-app>.onrender.com/health` returns
>    `{"status":"ok"}`.

## Assumptions

- In-memory storage is acceptable per the assignment's scope; ticket and user
  data does not persist across restarts/redeploys.
- Email is treated case-insensitively and trimmed of whitespace on
  register/login.
- Password minimum length of 6 characters is an arbitrary, simple validation
  rule (not specified in the brief but reasonable to include).
- JWTs are valid for 24 hours.
- No admin role, ticket assignment, or comments — out of scope per the brief.
