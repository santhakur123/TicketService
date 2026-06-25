# Ticket System (Golang Backend Intern Assignment)

A small backend service where users can register, log in, create tickets, view only
their own tickets, and update the status of their own tickets.

## Tech Stack & Design Choices

- **Language:** Go  (
  for method + path-parameter routing, e.g. `GET /tickets/{id}`).
- **Storage:** In-memory (a mutex-guarded map)
  
- **Auth:** JWT  implemented with the standard library
  rather
  than a third-party JWT package, so the project has **zero external
  dependencies** — no `go.sum`, nothing to fetch from the internet to build.
- **Passwords:** Never stored in plaintext. Stored as  where the hash
  is  rounds over `salt + password`. Verification uses a constant-time comparison.

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
├── password.go    # salted, iterated  password hashing
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
JWT_SECRET=mysecretkey ./ticket-system
curl http://localhost:8080/health
```

## Running with Docker

```bash
docker build -t ticket-system .
docker run -p 8080:8080 -e JWT_SECRET=mysecretkey ticket-system
curl http://localhost:8080/health
```

Expected response:
```json
{"status": "ok"}
```

## Environment Variables

See `.env.example`.





## Deployment

Deployed at: **https://ticketservice-kz87.onrender.com**
Public health check: **https://ticketservice-kz87.onrender.com/health**

##Some of the screenshots photo
<img width="200" height="100" alt="Screenshot 2026-06-25 202501" src="https://github.com/user-attachments/assets/225c2236-f7d8-4cf9-8874-063a3240a07c" />
<img  width="200" height="100" alt="Screenshot 2026-06-25 221007" src="https://github.com/user-attachments/assets/bea031f0-ef1a-40b6-856f-d65fee9a85aa" />
<img  width="200" height="100" alt="Screenshot 2026-06-25 221025" src="https://github.com/user-attachments/assets/9ee6c6fd-8501-4810-8314-90f900f1d642" />
<img  width="200" height="100" alt="Screenshot 2026-06-25 221223" src="https://github.com/user-attachments/assets/93ebb30d-84f7-4ad2-a7b6-c59e1fc3702c" />
<img  width="200" height="100" alt="Screenshot 2026-06-25 221344" src="https://github.com/user-attachments/assets/408fac4a-3010-48fd-b568-144e43b6298e" />
<img  width="200" height="100" alt="Screenshot 2026-06-25 221528" src="https://github.com/user-attachments/assets/b563e30a-545d-4b3d-8793-6ff3c2ee9a5b" />
<img  width="200" height="100" alt="Screenshot 2026-06-25 221548" src="https://github.com/user-attachments/assets/f4429df7-222d-4250-af86-960d457d8623" />
<img  width="200" height="100" alt="Screenshot 2026-06-25 221615" src="https://github.com/user-attachments/assets/603f6f2b-8f44-4455-a782-7ef98f105022" />










