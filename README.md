# Ingen One — Phase 1 MVP

Working vertical slice of the Ingen One platform (see [prd.md](prd.md)): Platform Core
(auth, tenancy, module entitlements), Sales (accounts/contacts/leads), Customer Service
(queues/cases), and Collaboration (1:1 chat over WebSocket), fronted by an installable
Nuxt 4 + shadcn-vue PWA.

## Stack

- **API** — Go + Fiber (`apps/api`): JWT auth, tenant-scoped REST, WebSocket chat gateway.
- **Frontend** — Nuxt 4 + TypeScript + Tailwind v4 + shadcn-vue + Pinia (`apps/frontend`),
  installable via `@vite-pwa/nuxt`. The app shell (collapsible sidebar, breadcrumb header,
  dark mode) follows the layout pattern from
  [nuxt-shadcn-dashboard](https://github.com/dianprata/nuxt-shadcn-dashboard), adapted so the
  sidebar is generated from `/me/manifest` instead of a static menu.
- **Database** — Postgres 16 (`db/migrations`), row-scoped by `organization_id`.

## Repository layout

```
apps/
  api/        Go + Fiber backend
    internal/
      auth/           JWT issuing + tenant-scoping middleware
      config/         env loading
      db/              pgx pool
      entitlement/     per-tenant module gate (404s disabled modules)
      modules/
        identity/      register, login, /me, /me/manifest
        sales/         accounts, contacts, leads
        service/       queues, cases
        collab/        1:1 chat hub + WebSocket handler
  frontend/   Nuxt 4 app (app/pages, app/stores, app/components/ui …)
db/migrations/ SQL schema (applied automatically by the db container)
docker-compose.yml   Postgres (+ hookup point for the API container)
```

## Running locally

### 1. Database

```bash
docker compose up -d db
```

Schema applies automatically on first boot via `db/migrations/0001_init.sql`.

### 2. API

```bash
cd apps/api
DATABASE_URL="postgresql://user:password@localhost:5432/ingen_one?sslmode=disable" \
JWT_SECRET="dev-secret-change-me" \
API_PORT=8080 \
go run .
```

`GET /health` should return `200 ok`. The root `.env` also has these values pre-filled;
`go run .` picks it up automatically if you don't override with shell env vars.

### 3. Frontend

```bash
cd apps/frontend
npm install   # first time only
npm run dev
```

Open http://localhost:3000 — it redirects to `/login`. Use **Create one** to self-register
a new organization; the new tenant is granted `sales`, `service`, and `collab` entitlements
and you land in a dashboard whose nav/widgets are generated entirely from
`GET /me/manifest`.

Frontend reads `NUXT_PUBLIC_API_BASE` / `NUXT_PUBLIC_WS_BASE` from `apps/frontend/.env`
(defaults to `localhost:8080`).

## What's implemented (Phase 1 slice)

- Email/password auth issuing JWTs; tenant + business unit scoping on every request.
- Self-serve org registration (creates org + HQ business unit + Admin role + default
  entitlements) — this is the "seed a tenant" path for the MVP; there is no invite-teammate
  flow yet (see Known gaps).
- **Module entitlement enforcement is live and data-driven**: disabling a module for an
  org makes its routes 404 immediately, no redeploy, no restart (per PRD §7.4). Verified by
  toggling `module_entitlements` directly and re-hitting the API.
- Sales: accounts / contacts / leads CRUD, tenant-isolated.
- Service: queues / cases CRUD, tenant-isolated.
- Collaboration: 1:1 chat over WebSocket (`/ws?token=<jwt>`), persisted to Postgres,
  delivered live to the recipient if connected, with history backfill on load.
- Frontend: shadcn-vue design system, entitlement-driven sidebar nav and dashboard,
  Sales/Service/Chat pages wired to the live API, installable PWA (manifest + service
  worker via `@vite-pwa/nuxt`).

All of the above was exercised end-to-end (curl + a live two-browser Playwright session)
against a real Postgres instance, not just at the unit level.

## Known gaps / next steps

- No invite-teammate flow — a second user in an existing org currently has to be inserted
  by hand. This is the natural next slice.
- RBAC is coarse (Admin role only, no per-entity privilege depth yet — §5.1 privilege-depth
  model is still TODO).
- No workflow engine, audit log, global search, or file attachments yet (later Phase 0/1
  epics per the PRD roadmap, §9.2 E8/E9/E11).
- SLA timers, routing, KB, and the live chat *widget* (customer-facing) are not built —
  only the internal agent-to-agent chat.
- Web Push (VAPID keys are already in `.env`) is not wired up; PWA installability works,
  but notifications are not yet sent.
- CORS is currently locked to `http://localhost:3000` in `apps/api/main.go` — update before
  deploying anywhere else.

## Local dev credentials

Running `go run ./apps/api/cmd/seed` (against the same `DATABASE_URL`) creates a "Dev Org"
tenant with `dev@ingen.local` / `password`. Otherwise just use **Create one** on `/register`
to spin up your own tenant.
