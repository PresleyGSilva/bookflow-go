# BookFlow Architecture

## Modules (`backend/internal/`)

- `auth` — register (creates org + admin user), login (returns orgs + access
  token, sets httpOnly refresh cookie), refresh, me. JWT access 15min, bcrypt.
- `professionals` — CRUD; fields: name, email, phone, timezone, active.
- `services` — CRUD; fields: name, duration_min, price_cents (integer), active.
- `availability` — computes slots for a professional/service/date from
  working-hours minus blocks minus existing appointments, stepped by
  `slot_step_min`.
- `bookings` — public no-auth booking by `code` (create, fetch, reschedule, cancel).
- `admin` — calendar view and reports (totals, per-professional, per-service,
  revenue_cents) plus settings (`cancel_min_hours`, `reschedule_min_hours`,
  `slot_step_min`).
- `middleware` — JWT auth, tenant (`X-Org-ID` / `X-User`) enforcement,
  JSON envelope, error mapping.
- `store` — pgx pool, migrations runner.

## Timezone (UTC)

- All timestamps stored as `timestamptz` in UTC.
- Each professional carries a `timezone` column (default `America/Sao_Paulo`).
- Conversion (local date + "HH:MM" ↔ UTC range) happens in the backend using
  the professional's timezone; the API returns `timezone` alongside slots and
  the frontend renders local times.

## Concurrency / exclusion

- `appointments` carries a GiST exclusion constraint on
  `(professional_id, tstzrange(starts_at, ends_at))` for non-cancelled rows,
  so two concurrent inserts on the same slot cannot both commit.
- Fallback: transactional `SELECT ... FOR UPDATE` on the professional's rows
  before insert. Loser receives `409` with code `SLOT_TAKEN`.

## Lead-time rules

- Settings (`admin/settings`): `cancel_min_hours` (default 2),
  `reschedule_min_hours` (default 4), `slot_step_min` (default 30).
- Cancel with less than `cancel_min_hours` before `starts_at` → `422 CANCEL_TOO_LATE`.
- Reschedule with less than `reschedule_min_hours` before `starts_at` →
  `422 RESCHEDULE_TOO_LATE`.

## Multi-tenancy via org_id + X-Org-ID

- Every tenant row carries `org_id`.
- Login returns `orgs[]`; frontend persists `orgs[0].id` in `localStorage`
  (`bookflow_org_id`) and sends `X-Org-ID` plus `X-User` on tenant calls.
- Backend resolves the org from `X-Org-ID`, verifies the user belongs to it,
  and scopes every query by `org_id`.

## Envelope

- Success: `{"data": ...}` (+ optional `"meta"`).
- Error: `{"error":{"code":"...","message":"..."}}`.
