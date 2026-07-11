IngenCore MVP (Phase 1) — Scope and Acceptance Checklist

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
- [x] Team management: invite-by-email (token link, 7-day expiry, single-use), accept-invite flow, role promote/demote, activate/deactivate, last-admin guard rail — Admin-only, server- and client-enforced, verified end-to-end in browser.
- [x] Privilege-depth RBAC (§5.1): per-entity (account/contact/lead/case) create/read/write/delete/assign privileges at none/user/bu/bu_children/org depth, most-permissive-role-wins, enforced on every Sales/Service CRUD endpoint. Verified: user-depth own-record restriction, org-depth full access, and the full BU hierarchy (bu vs bu_children vs unrelated sibling BU) via a purpose-built multi-BU test org.
- [x] Admin console for the above: `/team/roles` (create/delete custom roles, per-entity/action/depth privilege grid) and `/team/business-units` (create/rename/delete with cycle prevention and non-empty guards), plus role/BU assignment inline on the Team page. Verified end-to-end in browser: created a custom "Sales Rep" role restricted to leads only, created a custom business unit, assigned both to an invited member, and confirmed the member list reflected it after reload.
- [x] Service core (§5.3): SLA timers (per-org policy, computed due dates, pause on `waiting` status, idempotent respond action, live ok/warning/breached/met status), queue routing (manual/round_robin/capacity auto-assignment with per-queue membership), knowledge base (draft/publish, keyword-based case suggestions). Verified via 25 backend assertions (breach simulation, pause/resume timing, round-robin alternation, capacity load-balancing across queues, publish/unpublish RBAC) plus a full browser pass on Cases/Queues/KB.
- [x] Postgres Row-Level Security (§6/§7.2.3 tenancy NFR): a second, database-enforced tenant
  isolation layer on top of the existing application-layer organization_id filters, on all 17
  genuinely tenant-scoped tables. Required introducing a separate, unprivileged `app_role`
  Postgres role for the API's runtime connection (the docker-compose bootstrap role is a
  superuser, which RLS never restricts) and a per-request tenant-scoped transaction
  (`internal/db`) that `RequireAdmin`, `entitlement.Require`, and every handler now share.
  Verified directly against Postgres — `app_role` with no session context sees zero rows from
  a table holding multiple tenants' data; with the context set, exactly one tenant's rows —
  and end-to-end through the API across two fresh orgs (cross-tenant reads/writes 404) with a
  full regression pass confirming every existing feature still works unchanged.
- [x] Workflow engine (§5.1/E9): Admin-authored automations (`workflows` table) with a trigger (record created/updated on account/contact/lead/case, or a case SLA breach), a list of conditions (`equals`/`not_equals`/`changed`, all must match), and a sequence of actions (`update_field`, `assign_owner`, `notify`, `webhook`) — every match/run is logged to `workflow_runs`. Runs in-process off the request path (fire-and-forget goroutine per create/update; a 30s-poll background scanner for SLA breaches, deduped via per-case notified-at columns so a breach fires exactly once). Notifications land in-app (`notifications` table, unread count, mark-read) with a bell in the header. Admin-only `/automation/workflows` builder page (conditions/actions editor, run history, active/paused toggle) plus the notification bell were verified end-to-end in a real browser: created workflows for all 4 action types, confirmed condition matching (and non-matching records left untouched), confirmed the "changed" condition on an update trigger, confirmed an SLA breach (1-minute test policy) was caught by the scanner and fired exactly once, and confirmed the header bell's unread badge/dropdown/mark-read cycle.
- [x] Web Push (§5.11/§7.5): the workflow engine's `notify` action and offline 1:1 chat messages fan out to Web Push via a new `internal/push` package (`webpush-go`, VAPID-authenticated), storing subscriptions in a new RLS-protected `push_subscriptions` table. The PWA's service worker was rebuilt on `@vite-pwa/nuxt`'s `injectManifest` strategy (a custom `sw.ts`, not the previous fixed `generateSW` template) specifically so it can hold `push`/`notificationclick` listeners; this surfaced and fixed a real bug where the prior config's implicit navigation-fallback route would have thrown at service-worker-script-evaluation time on this Nitro node-server build (no static `index.html` exists to precache). Verified: subscribe/unsubscribe/vapid-key endpoints via direct API calls; delivery logic (send success, prune-on-410, log-on-other-error, log-on-network-error) against both a real push endpoint and a local mock server; the service worker registers and activates in a real browser with no console errors. Not independently verified: an actual `pushManager.subscribe()` call succeeding in-browser — Playwright's bundled open-source Chromium lacks the Google API keys real Chrome has for this, a known ecosystem limitation rather than an app bug.
- [x] Audit log (§5.1/E8): every create/update/delete on accounts, contacts, leads, and cases writes an `audit_log` row (actor, before/after JSON, timestamp) inside the same transaction as the mutation — an audit-write failure rolls back the whole request, so a write can never commit unrecorded. `app_role` has `UPDATE`/`DELETE` revoked on the table at the database level (append-only, enforced independent of application code, not just by convention). Admin-only `/audit` page: filter by entity, view actor/timestamp/action, and see the full before/after diff per entry. Verified end-to-end: all 4 entities × create/update/delete produced correct audit rows with accurate before/after values and actor attribution, entity filtering worked, org B's audit log was empty after only org A had activity (RLS + app-layer isolation), and a non-Admin member correctly got 403 from the endpoint.

Open Questions (resolved for this slice)
1) Environment: local Postgres via docker-compose for dev (`.env` already points at localhost:5432). Neon wiring is a config change (`DATABASE_URL`), not a code change — revisit before shared/staging deploys.
2) Frontend: Nuxt 4 + shadcn-vue (Reka UI) confirmed and scaffolded under `apps/frontend`.
3) Auth: email verification / Resend integration still deferred — email/password only, no verification step yet.
4) Tenancy model: self-serve registration creates one org + HQ business unit per signup; teams are grown via the invite flow.
5) CI: not yet set up — still open.

Next Steps After This Slice
- Global search and file attachments (Phase 0/1 epic E11, part of E8) remain — the workflow
  engine (E9), Postgres RLS (part of E4), Web Push (part of E10/E10b), and the audit log
  (part of E8) are all now built.
- The audit log currently covers account/contact/lead/case only (the same set the
  privilege-depth model already treats as "the" business records) — admin-config tables
  (roles, business units, workflows, queues, KB) are not audited yet, a reasonable follow-up.
- Sales pipeline (opportunities, stages, quotes) and 1:1 WebRTC calls remain from the earlier
  gap list — WebRTC calls are explicitly named in the PRD's Phase 1 exit criterion alongside
  push, which is now done.
- The customer-facing live chat widget (vs. the internal agent chat built here) is still Phase 2/3.
- Workflow engine known gaps: no NATS/event-bus (triggers are direct in-process calls from CRUD
  handlers, fine at Phase 1 scale but not horizontally scalable as-is); actions run with
  Admin-equivalent privilege rather than through the privilege-depth model (a workflow is
  authored by an Admin and trusted like a DB trigger); conditions compare fields as strings only
  (equals/not_equals/changed) — no numeric/date comparisons yet.
