# IngenCore — Phase 1 MVP

Working vertical slice of the IngenCore platform (see [prd.md](prd.md)): Platform Core
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
DATABASE_URL="postgresql://app_role:apppassword@localhost:5432/ingencore?sslmode=disable" \
JWT_SECRET="dev-secret-change-me" \
API_PORT=8080 \
go run .
```

Note the API connects as `app_role`, not the `user` superuser docker-compose bootstraps —
`app_role` is created by `db/migrations/0006_row_level_security.sql` and is what Row-Level
Security actually applies to (see "Security" below). Migrations themselves still run as the
superuser via `docker-entrypoint-initdb.d`/manual `psql`.

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
- Self-serve org registration (creates org + HQ business unit + Admin/Member roles + default
  entitlements) — this is the "seed a tenant" path for the MVP.
- **Team management**: Admins invite teammates by email (token-based invite link, 7-day
  expiry, single-use); invitee sets their own password via `/accept-invite`. Admins can
  promote/demote roles (Admin/Member) and activate/deactivate members, with a guard rail
  preventing an org from ever being left with zero active Admins. `RequireAdmin` middleware
  re-checks the DB on every request, so a role change takes effect immediately — no new
  token needed.
- **Module entitlement enforcement is live and data-driven**: disabling a module for an
  org makes its routes 404 immediately, no redeploy, no restart (per PRD §7.4). Verified by
  toggling `module_entitlements` directly and re-hitting the API.
- **Privilege-depth RBAC** (PRD §5.1): per-entity privileges (create/read/write/delete/assign)
  granted at a depth — none/user/bu/bu_children/org — on `account`, `contact`, `lead`, `case`.
  A user's effective depth for an action is the most permissive depth across all roles they
  hold (`internal/authz`). Admins default to org depth everywhere; Members default to
  org-depth create/read but user-depth write/delete/assign (can see everything, only touch
  what they own). List endpoints filter to the caller's scope; single-record endpoints
  return 404 (not 403) for records outside scope, consistent with how disabled modules are
  hidden rather than blocked. Verified end-to-end including the BU hierarchy: a parent-BU
  actor with `bu_children` depth can reach a child-BU record; `bu` depth (no children) and
  unrelated sibling BUs correctly cannot.
- Sales: accounts / contacts / leads CRUD, tenant-isolated, privilege-scoped.
- **Service core** (PRD §5.3):
  - **SLA timers** — one policy per org (first-response / resolution minutes, Admin-editable).
    Every case gets computed due dates at creation; a `waiting` status pauses both timers
    (tracked via `paused_seconds` + an in-progress `paused_at`, so the clock is genuinely
    frozen, not just hidden); a dedicated `respond` action marks first-response idempotently;
    status badges compute live as `ok` / `warning` (last 20% of the window) / `breached` /
    `met` / `met_late` — no polling job, it's derived at read time.
  - **Queue routing** — `manual` (default) / `round_robin` (cursor on the queue) / `capacity`
    (fewest open cases org-wide) auto-assignment on case creation, with per-queue membership
    managed from the UI (Admin-only).
  - **Knowledge base** — draft/published articles; a naive ILIKE keyword match suggests
    published articles from a case's subject (`GET /service/kb/suggest?case_id=`) — a
    placeholder until real search (Meilisearch) lands.
  - Queues themselves remain unscoped/Admin-managed (not owner-scoped records); cases are
    tenant-isolated and privilege-scoped like Sales.
- Collaboration: 1:1 chat over WebSocket (`/ws?token=<jwt>`), persisted to Postgres,
  delivered live to the recipient if connected, with history backfill on load.
- Frontend: shadcn-vue design system, entitlement-driven sidebar nav and dashboard,
  Sales/Service/Chat pages wired to the live API, installable PWA (manifest + service
  worker via `@vite-pwa/nuxt`).
- **Admin console for roles, privileges, and business units** (`/team/roles`,
  `/team/business-units`): create custom roles (default to no access on everything, then
  dial up per entity/action/depth from a grid), delete unused custom roles, and manage the
  BU tree (create/rename/delete with cycle prevention and non-empty guards). The Team page
  assigns any role and any business unit to a member inline. Everything here was previously
  SQL-only — this is the UI for the privilege-depth engine described above.
- **Workflow engine** (PRD §5.1/E9, `/automation/workflows`, Admin-only): trigger → conditions
  → actions automation on accounts/contacts/leads/cases. Triggers: record created, record
  updated, or (for cases) an SLA breach. Conditions (`equals`/`not_equals`/`changed`, all must
  match) gate a sequence of actions: `update_field`, `assign_owner`, `notify` (in-app, shown via
  the bell in the header — unread count, dropdown, mark-read), and `webhook` (POSTs
  `{entity, record_id, fields}`). Every match is logged to a per-workflow run history (status +
  detail) visible from the admin page. Created/updated triggers fire from the Sales/Service
  handlers as a fire-and-forget goroutine (10s timeout, decoupled from the request's own
  context since Fiber recycles it after the handler returns); SLA breaches are caught by a
  30-second background scanner that dedupes via a per-case notified-at column so a breach
  fires exactly once. This is what turns an SLA breach from a red badge into an actual
  notification/reassignment/webhook call.
- **Web Push** (PRD §5.11/§7.5): the workflow engine's `notify` action and offline 1:1 chat
  messages now also fan out to Web Push, so an assignment, SLA breach, or DM reaches the
  user even with the PWA closed. A dedicated, unprivileged Postgres role isn't needed here
  (push subscriptions are just another RLS-protected tenant table), but delivery is real:
  `internal/push` sends VAPID-authenticated requests via `webpush-go`, pruning a subscription
  the moment its push service reports it gone (404/410) and logging any other failure —
  verified against both a live push endpoint and a local mock server that can be told to
  return any status. The PWA's service worker (`app/service-worker/sw.ts`, built via
  `@vite-pwa/nuxt`'s `injectManifest` strategy so it can hold custom code) handles the `push`
  and `notificationclick` events; a toggle in the header's notification bell requests
  permission and registers the subscription.
- **Audit log** (PRD §5.1/E8, and named in the §6 NFRs: "audit trail immutable
  (append-only)"): every create/update/delete on accounts, contacts, leads, and cases — the
  same four entities the privilege-depth model already treats as "the" business records —
  writes an `audit_log` row with actor, before/after JSON, and timestamp, in the same
  transaction as the mutation itself (`internal/audit`), so a write can never commit without
  also being recorded. `app_role` only has `SELECT`/`INSERT` on the table — `UPDATE`/`DELETE`
  are revoked at the database level, so even a compromised app credential can't rewrite
  history. Admin-only `/audit` page: filter by entity, see actor/timestamp/action per row, and
  view the full before/after diff. Verified end-to-end: all 4 entities × create/update/delete
  logged with correct before/after values and actor attribution, entity filtering, cross-tenant
  isolation (org B sees zero of org A's entries), and non-Admins correctly forbidden (403).

All of the above was exercised end-to-end (curl + a live two-browser Playwright session)
against a real Postgres instance, not just at the unit level.

## Security

**Postgres Row-Level Security** (PRD §6/§7.2.3 tenancy NFR: "never rely on RLS alone or app
checks alone — both"): every tenant-scoped table (17 in total — accounts, contacts, leads,
cases, queues, messages, workflows, notifications, business_units, roles, role_privileges,
user_roles, etc.) has `ENABLE`+`FORCE ROW LEVEL SECURITY` with a policy comparing
`organization_id` against a Postgres session variable (`app.org_id`) that fails closed (zero
rows) if never set. This is a second, database-enforced isolation layer independent of the
existing application-layer `organization_id` filters — if an app-layer filter is ever missing
or buggy, RLS still blocks the cross-tenant read/write rather than leaking it.

This required two things beyond the policies themselves: (1) a dedicated, unprivileged
`app_role` Postgres role for the API's runtime connection — the docker-compose bootstrap
`user` role is a **superuser**, and RLS has no effect on superusers/BYPASSRLS roles no matter
how the policies are written; migrations still run as the superuser, only the running API
connects as `app_role`; (2) every authenticated request now runs inside one transaction
(`internal/db` `Middleware`/`Tx`) with `app.org_id` set from the JWT, so `RequireAdmin`,
`entitlement.Require`, and every handler's queries all see the same tenant context. A handful
of code paths that legitimately need cross-tenant access before an org context exists —
login-by-email, invite-by-token/accept-invite, org registration itself, and the SLA breach
scanner (a background job with no per-request context, refactored to loop org-by-org via
`db.WithOrgTx`) — are the documented exceptions, either left off RLS (`organizations`,
`users`, `sessions`, `invites`) or given their own short-lived scoped transaction once the org
id becomes known mid-handler.

Verified directly against Postgres (not just through the API): connecting as `app_role` via
psql with no session context returns zero rows from a table holding data across multiple
tenants; setting the session variable to one tenant's org id returns exactly that tenant's
rows. Also verified end-to-end through the API: two fresh orgs, cross-tenant reads/writes on
accounts and role privileges correctly 404, and every existing feature (auth, team/RBAC,
Sales/Service CRUD, the SLA breach scanner, the workflow engine and notifications, 1:1 chat)
still passes with the new tenant-scoped transaction plumbing in place.

A full review pass (dependency scan + multi-angle code review + manual verification) found
and fixed:
- **Dependency vulnerabilities**: `pgx` (SQL-sanitization bug, GO-2026-5004) and `fiber`
  (3 DoS/crash bugs) were behind their patched versions and reachable from our code paths
  per `govulncheck`; bumped to `pgx@v5.10.0` / `fiber@v2.52.14`. Go toolchain pinned to
  `1.26.5` in `go.mod` for a stdlib TLS fix. `govulncheck` and `npm audit` both report zero
  reachable vulnerabilities as of this pass.
- **Privilege-scope bypass**: `GET /service/kb/suggest` looked up a case by org only, not by
  the caller's privilege depth — a `user`-depth reader could learn another user's case
  subject through it even though `GET /service/cases/:id` on the same case correctly 404'd.
  Fixed to use the same `authz.Guard`/`ScopedWhere` scoping as every other case endpoint.
- **Cross-tenant reference**: `POST /team/invites` accepted a `business_unit_id` without
  checking it belonged to the inviter's own org (unlike `PATCH /team/members/:id`, which
  already did). Fixed to validate via the same `buExists` check.
- **Correctness bug in the privilege-depth filter**: `authz.ScopedWhere`'s placeholder index
  was miscounted in `UpdateContact`, `UpdateLead`, and `UpdateCase` (both its read and write
  queries) — any non-org-depth role (e.g. the default Member role) got a SQL parameter-count
  error and a 500 trying to edit even their own contacts/leads/cases. Fixed all four; every
  other `ScopedWhere` call site was audited and confirmed correct.

## Known gaps / next steps

- No audit log, global search, or file attachments yet (later Phase 0/1 epics per the PRD
  roadmap, §9.2 E8/E11).
- Workflow engine (E9) runs in-process, triggered by direct calls from the CRUD handlers —
  there's no NATS/event-bus layer yet (fine at Phase 1 scale, not horizontally scalable
  as-is). Workflow actions run with Admin-equivalent privilege rather than through the
  privilege-depth model — a workflow is authored by an Admin and trusted the same way a
  database trigger would be. Conditions compare fields as strings only (equals/not_equals/
  changed) — no numeric/date comparisons yet.
- SLA policy is one-per-org (no per-priority or per-queue matrices yet); skills-based routing
  and the live chat *widget* (customer-facing) are Phase 2/3 per the PRD and not built —
  only internal agent-to-agent chat and round-robin/capacity queue routing.
- Web Push delivery has one real-environment gap: creating a browser Push subscription
  (`pushManager.subscribe()`) needs a browser with real push-service credentials — Playwright's
  bundled open-source Chromium lacks Google's proprietary API keys and always fails this one
  call with "permission denied," a well-known limitation of open-source Chromium builds, not
  an app bug. Everything else was verified in a real browser and against the real API: the
  service worker registers/activates, subscribe/unsubscribe/vapid-key endpoints work, and
  delivery (including pruning a 410 subscription and logging other failures) was verified
  against both a real push endpoint and a local mock server.
- CORS is currently locked to `http://localhost:3000` in `apps/api/main.go` — update before
  deploying anywhere else.
- Invite links are surfaced in the Team UI for the Admin to copy/share manually — the
  Resend API key in `.env` is a placeholder and email sending isn't wired up yet.

## Local dev credentials

Running `go run ./apps/api/cmd/seed` (against the same `DATABASE_URL`) creates a "Dev Org"
tenant with `dev@ingencore.local` / `password`. Otherwise just use **Create one** on `/register`
to spin up your own tenant.
