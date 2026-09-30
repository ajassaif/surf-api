# surf-api

A small surf lesson booking API for a surf school in Varkala, built with [GoFr](https://gofr.dev) and SQLite.

## Run

```bash
go mod tidy
go run .
```

GoFr reads `configs/.env`, creates `surf.db`, runs the migrations, and starts on `:8000`.

## Endpoints

| Method | Path | What it does |
|---|---|---|
| GET | `/instructors` | List instructors |
| GET | `/bookings?date=YYYY-MM-DD` | List bookings (date filter optional) |
| POST | `/bookings` | Book a lesson: `{"guest_name","instructor_id","date","slot"}` (slot: `morning`/`evening`) |
| DELETE | `/bookings/{id}` | Cancel a booking |
| GET | `/.well-known/health` | Health check (built into GoFr) |

Double-booking an instructor for the same date and slot returns `409 Conflict`.

## Test

With the server running: `./smoke_test.sh`. CI runs the same script on every push.
