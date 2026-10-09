# BookFlow — Appointment Scheduling (Go + Next.js + Postgres)

BookFlow is a multi-tenant appointment scheduling system: professionals, services,
working hours, blocks, customers, public booking links, and admin
calendar/reports. Backend is a modular Go monolith (Chi + pgx); frontend is
Next.js; Postgres 16 is the store.

## Quickstart

```bash
cp .env.example .env
docker compose up --build
```

- API: http://localhost:8085 (`GET /healthz`, `GET /openapi.yaml`)
- Web: http://localhost:3005
- Postgres (host): localhost:5438

Demo seed login: `admin@cortefino.com` / `admin123` (org "Barbearia Corte Fino",
3 professionals, 4 services, Mon–Sat working hours). Migrations run from
`backend/migrations/` (via the compose `migrate` service, version-guarded, or
psql); seed lives in `0002_seed.sql` (idempotent).

## Endpoints (`/api/v1` prefix)

- Auth: `POST /auth/register`, `POST /auth/login`, `POST /auth/refresh`, `GET /auth/me`
- Professionals: `GET/POST /professionals`, `GET/PATCH/DELETE /professionals/{id}`
- Services: `GET/POST /services`, `GET/PATCH/DELETE /services/{id}`
- Working hours: `GET/POST /working-hours?professional_id=`, `DELETE /working-hours/{id}`
- Blocks: `GET/POST /blocks?professional_id=`, `DELETE /blocks/{id}`
- Customers: `GET/POST /customers`, `GET/PATCH /customers/{id}`
- Availability: `GET /availability?professional_id=&service_id=&date=YYYY-MM-DD`
- Public catalog (no auth, powers `/book`): `GET /public/orgs`,
  `GET /public/services?org_id=`, `GET /public/professionals?org_id=`,
  `GET /public/availability?professional_id=&service_id=&date=YYYY-MM-DD`
- Public bookings (no auth): `POST /bookings/public`, `GET /bookings/public/:code`,
  `POST /bookings/public/:code/reschedule`, `POST /bookings/public/:code/cancel`
- Appointments (auth): `GET /appointments`, `POST /appointments`,
  `POST /appointments/:id/reschedule`, `POST /appointments/:id/cancel`
- Admin: `GET /admin/calendar`, `GET /admin/reports`, `GET/PATCH /admin/settings`
- Misc: `GET /healthz`, `GET /openapi.yaml`

Success: `{"data": ...}` (+ optional `"meta"`). Errors:
`{"error":{"code":"...","message":"..."}}`.

## Business rules

- Money is stored in integer **cents** (never float).
- Times are stored as UTC `timestamptz`; each professional has a `timezone`
  (default `America/Sao_Paulo`); conversion happens in the backend.
- Slot conflicts: only one booking wins per professional/time range
  (exclusion constraint or transactional lock) — loser gets `409 SLOT_TAKEN`.
- Cancel requires >= `cancel_min_hours` (default 2h) notice, else `422 CANCEL_TOO_LATE`.
- Reschedule requires >= `reschedule_min_hours` (default 4h) notice, else
  `422 RESCHEDULE_TOO_LATE`.
- Public access uses the appointment `code` (e.g. `BF-XXXXXX`).
- Tenant scoping: login returns `orgs[]`; the frontend persists `orgs[0].id`
  in `localStorage` (`bookflow_org_id`) and sends `X-Org-ID` + `X-User`.

## Screenshots (real, local stack)

![Landing](docs/screenshots/01-landing.png)
![Public booking — slots](docs/screenshots/02-book-slots.png)
![Public booking — confirm](docs/screenshots/03-book-confirm.png)
![Booking confirmed](docs/screenshots/04-book-done.png)
![Login](docs/screenshots/05-login.png)
![Admin calendar](docs/screenshots/06-admin-calendar.png)
![Admin reports](docs/screenshots/07-admin-reports.png)

## Layout

- `backend/` — Go API, migrations in `backend/migrations/`, spec in `backend/openapi.yaml`
- `web/` — Next.js frontend
- `docs/ARCHITECTURE.md` — modules, timezone, concurrency, multi-tenancy

See `README.pt-BR.md` for the Portuguese version. License: MIT.
