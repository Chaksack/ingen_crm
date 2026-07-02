Ingen One MVP (Phase 1) — Scope and Acceptance Checklist

Scope Overview (mapped from prd.md)
- Platform Core
  - Identity: email/password auth (dev-only), sessions (JWT), basic device/session tracking.
  - Multi-tenancy: organizations, business units, users; row-level isolation plan.
  - RBAC: roles and user-role mapping; coarse-grained checks in API middleware (later).
  - Module entitlements: per-tenant activation of sales, service, collab (Phase 1 focus).
  - API: REST baseline + WebSocket for chat; health endpoint.
- Sales (CRM lite)
  - Entities: accounts, contacts, leads (CRUD minimal); opportunities deferred.
- Customer Service (lite)
  - Entities: queues (placeholder), cases (CRUD minimal).
- Collaboration (Phase 1)
  - 1:1 direct messages via WebSocket; message persistence.
- PWA (Phase 1)
  - Installable app via @vite-pwa/nuxt; offline shell; push later.

Non-Functional (initial targets)
- Repo structure to support modularity; docker-compose for local dev; migrations in SQL.
- Secrets policy: .env.example only; no real secrets in VCS.

Acceptance Checklist (MVP skeleton)
- [x] Repo layout: apps/api, apps/frontend, db/migrations, docs present.
- [x] API builds and serves /health; configurable via env; containerized.
- [x] Database migrations apply cleanly to Postgres 16.
- [x] Tables exist for organizations, business_units, users, roles, user_roles, module_entitlements, sessions, accounts, contacts, leads, queues, cases, messages.
- [x] docker-compose up starts Postgres; API runs locally against it, /health returns 200 OK.
- [x] .env has real local values (not committed as an example — see README's "Running locally").
- [x] README documents run steps.
- [x] Auth endpoints (register/login) + JWT middleware with tenant scoping.
- [x] CRUD endpoints for Accounts, Contacts, Leads, Cases, Queues.
- [x] Module entitlements enforced server-side: disabled module -> 404, verified live.
- [x] WebSocket server for 1:1 DMs with message persistence and history backfill.
- [x] Frontend shell (Nuxt 4 + shadcn-vue) + entitlement-driven nav/dashboard + PWA manifest/service worker.
- [x] Login, register, Sales/Service list+create screens, and a chat pane — all wired to the live API and verified in a real (Playwright-driven) browser session, including live message delivery across two logged-in sessions.

Open Questions (resolved for this slice)
1) Environment: local Postgres via docker-compose for dev (`.env` already points at localhost:5432). Neon wiring is a config change (`DATABASE_URL`), not a code change — revisit before shared/staging deploys.
2) Frontend: Nuxt 4 + shadcn-vue (Reka UI) confirmed and scaffolded under `apps/frontend`.
3) Auth: email verification / Resend integration still deferred — email/password only, no verification step yet.
4) Tenancy model: self-serve registration creates one org + HQ business unit per signup; no invite-teammate flow yet (next slice).
5) CI: not yet set up — still open.

Next Steps After This Slice
- Invite-teammate flow (currently the only way to add a second user to an org is a direct DB insert).
- Per-entity privilege depth (user/BU/BU+children/org) — currently a single coarse Admin role.
- Workflow engine, audit log, global search, file attachments (Phase 0/1 epics E8/E9/E11).
- Wire up Web Push (VAPID keys already provisioned in `.env`) for assignment/SLA/message notifications.
- SLA timers + routing for Service, and the customer-facing live chat widget (vs. the internal agent chat built here).
