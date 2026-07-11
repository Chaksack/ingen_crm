# Product Requirements Document
## "IngenCore" — Unified Business Applications Platform
### A Dynamics 365-class ERP/CRM suite with native real-time collaboration

**Version:** 1.1 | **Date:** July 2026 | **Owner:** Andrew | **Status:** Draft for review

**v1.1 changes:** shadcn-vue design system confirmed; per-tenant module entitlement architecture added (§5.10, §7.4); installable PWA with real-time web push added as a Phase 1 requirement.

---

## 1. Executive Summary

IngenCore is a self-hosted, multi-tenant business applications platform delivering the core capabilities of Microsoft Dynamics 365 — Sales, Customer Service, Finance, Supply Chain/Operations, HR, and Project Operations — combined with a **native collaboration suite** (persistent chat, channels, meeting links, and audio/video calling) built into the platform rather than bolted on via Microsoft Teams.

The platform is built on a modern, cost-efficient stack: **Nuxt (TypeScript) + shadcn-vue** frontend, **Go (GoFiber)** backend services, and **Neon** (serverless Postgres) as the primary datastore, with WebRTC infrastructure for real-time media.

The platform is **modularly composed per tenant**: each organization is entitled to a set of modules (e.g., Service + Collaboration only, or the full ERP), and its navigation, dashboards, and API surface reflect exactly that set. The app ships as an **installable PWA** with real-time push notifications, so agents and managers get a native-app experience on desktop and mobile without app stores.

The strategic driver: eliminate per-seat Dynamics 365 + Teams licensing costs for a multi-client BPO/CX operation, own the data isolation model end-to-end, and productize the platform as a sellable asset.

**Delivery model is phased.** Full Dynamics parity is a multi-year effort; this PRD defines the complete vision but commits to an MVP (Phase 1) centered on CRM, Customer Service, and the collaboration suite — the modules that directly unblock BPO operations — before Finance and Supply Chain.

---

## 2. Background & Problem Statement

- Dynamics 365 licensing (Customer Service Enterprise, Contact Center, Sales) plus Teams scales linearly with seats. For a growing BPO, per-agent-per-month costs become a dominant operating expense and margins compress as headcount grows.
- Multi-client data isolation in Dynamics requires careful business-unit/security-role architecture; owning the platform means owning the isolation model natively.
- Context-switching between the CRM and a separate communications tool (Teams) costs agent productivity. Embedding chat, calls, and meetings directly beside the case/deal record removes that switch.
- No mainstream open alternative combines ERP + CRM + native calling in one coherent, multi-tenant platform.

**Problem statement:** Service businesses need an integrated operations platform (sell, serve, bill, fulfil, staff) with built-in team and customer communication, without per-seat licensing of two separate ecosystems.

## 3. Goals & Non-Goals

### Goals
1. Replace Dynamics 365 CE (Sales + Customer Service) for internal BPO use by end of Phase 1.
2. Native collaboration: 1:1 and channel chat, presence, audio/video calls, and shareable meeting links (guests can join from a browser without an account).
3. First-class multi-tenancy: one deployment, many client organizations, strict row-level data isolation.
4. Finance and Operations parity sufficient to run a real services business (GL, AR/AP, invoicing, procurement, inventory) by Phase 3.
5. Extensibility: custom entities, custom fields, workflow automation, and a REST/webhook API — the "Power Platform-lite" layer.
6. **Per-tenant modularity:** modules are entitlement-gated per organization; a tenant's dashboard, navigation, and API expose only its subscribed modules — enabling tiered packaging (Service-only, CRM+Service, full ERP).
7. **Installable PWA:** users can install the app on desktop/Android/iOS and receive real-time push notifications (assignments, SLA warnings, messages, incoming calls) even when the app is closed.
8. Total infrastructure cost that undercuts equivalent Dynamics + Teams licensing at ≥25 seats.

### Non-Goals (explicitly out of scope for v1.x)
- PSTN/telephony carrier integration (dial real phone numbers) — Phase 4+ via SIP trunk provider.
- On-premise/air-gapped deployment packaging.
- Marketplace/app-store for third-party extensions.
- AI copilot features (deferred; architecture must not preclude them).
- Native app-store apps (the installable PWA is the mobile experience until Phase 4).
- Feature-for-feature parity with every Dynamics module (e.g., Field Service IoT, Mixed Reality guides).

## 4. Users & Personas

| Persona | Description | Primary modules |
|---|---|---|
| **Agent** | BPO front-line agent handling cases/chats for one or more client accounts | Customer Service, Chat, Calls |
| **Team Lead / Supervisor** | Monitors queues, SLAs, agent performance; joins escalation calls | Customer Service, Dashboards, Meetings |
| **Sales Rep / AE** | Manages leads → opportunities → quotes → orders | Sales, Chat, Meetings |
| **Finance Officer** | Invoicing, AR/AP, GL, reporting, period close | Finance |
| **Ops Manager** | Procurement, vendors, inventory, fulfilment | Supply Chain/Ops |
| **HR / People Ops** | Employee records, leave, onboarding | HR |
| **Org Admin** | Tenant configuration, users, roles, security, customization | Admin/Platform |
| **Platform Owner (Andrew)** | Cross-tenant super-admin, billing, provisioning | Platform console |
| **External Guest** | Client stakeholder joining a meeting via link; customer on a support chat | Meetings, Customer portal |

## 5. Product Scope Overview

The platform is organized as one shell application with modules (mirroring Dynamics "apps") on a shared data platform (mirroring Dataverse).

### 5.1 Platform Core (the "Dataverse equivalent") — Phase 1
- **Identity & access:** email/password + TOTP 2FA, SSO (OIDC/SAML) later; session management; device management.
- **Multi-tenancy:** organizations → business units (hierarchical) → teams → users. Every record owned by a user/team and scoped to a BU. Enforced with Postgres Row-Level Security in Neon plus application-layer checks.
- **RBAC:** security roles with per-entity privileges (create/read/write/delete/append/assign/share) at depth levels: user / BU / BU+children / org — the Dynamics privilege-depth model.
- **Metadata-driven entities:** system entities plus admin-defined custom entities and custom fields (typed columns, option sets, lookups); auto-generated forms and views.
- **Views, forms, dashboards:** configurable grids, saved queries, form designer (JSON-schema-driven), dashboard designer with charts.
- **Workflow/automation engine:** trigger (record created/updated/status change, schedule) → conditions → actions (update record, assign, notify, webhook, send email). The "Power Automate-lite."
- **Audit log:** entity-level auditing of all writes with actor, timestamp, before/after values.
- **Files & notes:** attachments on any record (object storage), timeline/activity feed per record.
- **Search:** global full-text search across permitted records.
- **Module registry & entitlements:** every module registers a manifest (entities, nav items, dashboards, permissions, workflows); tenant subscriptions gate which modules are active per organization (see 5.10).
- **Notifications:** in-app, email, and real-time web push (works when the PWA is closed) with per-user preferences and quiet hours.
- **API:** versioned REST API + webhooks; API keys and OAuth clients per tenant.

### 5.2 Sales (CRM) — Phase 1
- Leads: capture (manual, web form, CSV import), qualification, disqualification with reason, lead → contact/account/opportunity conversion.
- Accounts & Contacts with relationship hierarchy, deduplication rules.
- Opportunities: pipeline stages (customizable business process flow), probability, estimated/actual revenue, competitors, stakeholders.
- Products & price lists: units, currencies, discount lists.
- Quotes → Orders → Invoices (invoice handoff to Finance in Phase 3; standalone PDF invoice in Phase 1).
- Activities: tasks, appointments, emails (send/receive via SMTP/IMAP or Gmail/Microsoft OAuth), phone call logging — all on the record timeline.
- Sales dashboards: pipeline by stage, win/loss, forecast, activity leaderboards.
- Goals/quotas per rep and team.

### 5.3 Customer Service — Phase 1 (BPO-critical)
- Cases: subjects, priorities, origins (email, chat, portal, manual), status reasons, parent/child cases, merge.
- **Queues & routing:** per-client queues; round-robin and capacity-based auto-assignment; skills-based routing (Phase 2).
- **SLAs:** first-response and resolution timers with pause conditions, warning/breach actions, visible countdown on the case form.
- Entitlements: support contracts per client with case allotments.
- **Omnichannel inbox:** unified agent workspace — email-to-case, **live chat widget** (embeddable on client websites), and internal escalation. Voice channel arrives with the calling stack (5.8).
- Knowledge base: articles with versioning, approval flow, article suggestions on cases, public/portal visibility flags.
- Customer portal: clients' end-customers can raise and track tickets, read KB articles.
- Supervisor console: real-time queue depth, agent presence/capacity, SLA risk board, barge/whisper on chats (Phase 2 for calls).
- CSAT: post-resolution survey (email/chat), scores on dashboards.

### 5.4 Finance — Phase 3
- General Ledger: configurable chart of accounts, dimensions (department, client, project), journals, recurring journals, period open/close, fiscal calendars.
- Accounts Receivable: customer invoices (from Sales orders or manual), credit notes, payment recording, aging reports, dunning/reminder automation.
- Accounts Payable: vendor bills, approval workflow, payment runs, aging.
- Multi-currency: transaction + base currency, daily rates, realized/unrealized gain-loss.
- Tax: configurable tax codes (VAT/levies — UK VAT and Ghana VAT/NHIL/GETFund presets).
- Banking: bank accounts, manual statement import (CSV/OFX), reconciliation workbench.
- Expense management: employee expense reports with receipt capture and approval flow.
- Budgets: budget vs. actuals by dimension.
- Financial reporting: P&L, balance sheet, trial balance, cash-flow (indirect), by dimension; export to XLSX/PDF.
- **Explicitly deferred:** fixed-asset depreciation automation (basic register only), payroll (integrate, don't build), statutory e-filing.

### 5.5 Supply Chain / Operations — Phase 3
- Product/item master: stocked, non-stocked, services; SKUs, barcodes, unit conversions.
- Inventory: multi-warehouse, locations/bins, on-hand tracking, movements, adjustments, transfers, stock counts, FIFO costing (weighted-average Phase 4).
- Procurement: purchase requisitions → approvals → purchase orders → receiving → three-way match with AP bills; vendor management and vendor price lists.
- Sales order fulfilment: reservations, picking, packing, shipment, backorders.
- Reorder policies: min/max with suggested purchase orders.
- **Explicitly deferred:** manufacturing (BOM/production orders), demand forecasting, warehouse mobile scanning apps.

### 5.6 Project Operations — Phase 2
- Projects with phases/tasks, assignments, Gantt and board views.
- Time tracking: timesheets with approval, billable/non-billable.
- Resource utilization view.
- Project billing: fixed price milestones or time-and-materials → invoices (Finance handoff in Phase 3).

### 5.7 HR (lite) — Phase 2
- Employee records (linked to platform users), org chart, employment history, documents.
- Leave management: policies, accruals, requests, approvals, team calendar.
- Onboarding/offboarding checklists.
- **Deferred:** payroll, performance reviews, recruiting.

### 5.8 Collaboration Suite ("native Teams") — Phase 1–2
The flagship differentiator. Built in, not integrated.

**Chat & channels — Phase 1**
- 1:1 and group direct messages; public/private channels per organization and per team.
- Threads, @mentions, reactions, pinned messages, file sharing, link previews, message edit/delete with history, read receipts, typing indicators.
- Presence: available/busy/in-a-call/away, driven by activity and call state.
- **Record-linked chats:** every case, opportunity, or project can have a linked discussion visible on the record timeline — chat in the context of work.
- Full-text message search; retention policies per org.
- Delivery over WebSocket with offline queueing and push notifications.

**Calling — Phase 1 (1:1), Phase 2 (group)**
- 1:1 audio and video calls from any chat or from a contact/user profile, WebRTC peer-to-peer with TURN relay fallback.
- Group calls up to 50 participants via SFU (selective forwarding unit).
- In-call: mute, camera toggle, screen sharing, device selection, network-quality indicator.
- Call history and missed-call notifications integrated with chat and the activity timeline (calls with customers log to the record).

**Meetings — Phase 2**
- **Meeting links:** create a meeting (ad-hoc or scheduled) and get a shareable URL; **guests join from the browser with no account** — name entry + lobby admission by the host.
- Scheduling: calendar within the platform, availability view, ICS email invites so external parties get calendar entries; recurring meetings.
- In-meeting: participant roster, hand raise, in-meeting chat, host controls (mute all, remove, lock), screen share with presenter switching.
- Recording to object storage with access control (Phase 3); waiting room/lobby; join before host toggle.

**Voice channel for Customer Service — Phase 3**
- WebRTC-based customer callback ("call us" button on portal/widget) routed into service queues with the same SLA/routing engine; supervisor barge/whisper. PSTN via SIP trunk is Phase 4.

### 5.9 Marketing (lite) — Phase 4
- Segments (dynamic lists from CRM data), email campaigns with template editor, send tracking (open/click), simple drip journeys, web forms → leads. Full marketing automation is out of scope.

### 5.10 Per-Tenant Modular Composition — Phase 0–1 (foundational)
The platform is one deployment, but every tenant experiences only the modules they are entitled to. This is a first-class platform capability, not UI hiding.

- **Module manifests:** each module (sales, service, finance, scm, projects, hr, collab, marketing) declares its entities, navigation sections, dashboard widgets, security privileges, workflow triggers, and API routes in a registered manifest.
- **Tenant entitlements:** each organization has a subscription record listing enabled modules (with optional seat counts and feature flags per module). The platform console (super-admin) provisions and changes entitlements at runtime — no redeploy.
- **BU-level scoping:** within a tenant, business units can be restricted to a module subset — in the BPO model, agents assigned to client A see only the modules and queues in client A's engagement.
- **API enforcement:** middleware rejects any request touching a disabled module's entities or routes; disabled modules emit no events, run no workflows, and are excluded from search indexing for that tenant. Enforcement is server-side; the UI merely reflects it.
- **Dynamic dashboard & navigation:** on login the Nuxt shell fetches the effective manifest (tenant entitlements ∩ BU scope ∩ user role) and renders navigation, home dashboard widgets, and global-search scope from it. Role-appropriate default dashboards per module (agent → queues/SLA; sales rep → pipeline; finance → cash/aging), plus user-customizable widget layouts saved per user.
- **Packaging:** entitlement tiers double as commercial SKUs (e.g., Service Desk = service + collab; Sales Suite = sales + collab; Business = + projects/hr; Enterprise = full ERP). Usage metering per module feeds billing.

### 5.11 Progressive Web App & Push Notifications — Phase 1
- **Installable:** web app manifest + service worker; install prompts on desktop (Chrome/Edge), Android, and iOS (Add to Home Screen); standalone window, app icon, splash screen, theme color.
- **Real-time push:** Web Push (VAPID) delivers notifications when the app is backgrounded or closed — case assignments, SLA warning/breach, @mentions and DMs, incoming call rings, approval requests, meeting reminders. Tapping a notification deep-links to the record/chat/call.
- **iOS note:** iOS supports Web Push only for installed (Home Screen) PWAs on iOS 16.4+; onboarding flow must coach iPhone users to install.
- **Offline behavior:** app shell, navigation, and recently viewed records cached (stale-while-revalidate); message drafts and unsent chat messages queued and synced on reconnect. Full offline editing of business records is out of scope for v1.
- **Update flow:** service-worker driven "new version available — refresh" prompt; no forced reloads mid-call.
- **Badging:** unread counts on the app icon where supported (Badging API).

## 6. Non-Functional Requirements

| Area | Requirement |
|---|---|
| Tenancy & isolation | Row-level isolation per organization enforced by Postgres RLS + application guards; per-BU isolation within a tenant (the BPO multi-client model). Cross-tenant access impossible by construction; automated isolation tests in CI. |
| Security | OWASP ASVS L2; encryption in transit (TLS 1.3) and at rest; secrets in a vault; 2FA; rate limiting; CSP; dependency scanning; audit trail immutable (append-only). |
| Compliance | GDPR + UK GDPR + Ghana Data Protection Act: data export, right-to-erasure workflows, consent tracking, EU/UK-region hosting option (Neon region pinning). |
| Performance | P95 API < 300 ms; grid loads (50 rows) < 1 s; chat message delivery < 500 ms end-to-end; call setup < 3 s. |
| Scale targets | v1: 500 concurrent users/tenant, 50 tenants, 10 M records/entity, 200 concurrent calls. Architecture headroom to 10×. |
| Availability | 99.9 % for app and chat; graceful degradation — if media servers fail, chat still works; if workflow engine fails, CRUD still works. |
| Data | Point-in-time recovery via Neon; nightly logical backups to independent storage; RPO ≤ 5 min, RTO ≤ 4 h. |
| PWA/Push | Lighthouse PWA installability checks pass; push delivery P95 < 5 s app-closed; push subscriptions pruned on failure; icon badging where supported. |
| Accessibility | WCAG 2.1 AA for agent-facing surfaces. |
| i18n | English at launch; string externalization from day one; multi-currency and timezone-safe datetimes throughout. |
| Observability | Structured logs, metrics, tracing (OpenTelemetry); per-tenant usage metering (seats, storage, call minutes). |

## 7. Technical Architecture

### 7.1 Stack summary

| Layer | Technology | Notes |
|---|---|---|
| Frontend | **Nuxt 3+ (TypeScript)**, Pinia, TailwindCSS, **shadcn-vue** (Reka UI) | SSR for portal/public pages, SPA mode for the app shell; design system = shadcn-vue components themed per tenant; TanStack Table for grids |
| PWA | @vite-pwa/nuxt (Workbox service worker), Web Push (VAPID) | Installable app, offline shell caching, push subscriptions stored per user/device; Go service sends pushes via webpush library |
| API | **Go + GoFiber** | Modular monolith first (module-per-package), split to services only when scale demands |
| Realtime gateway | Go WebSocket service (or Centrifugo) | Chat delivery, presence, live dashboards, call signaling |
| Media | **LiveKit (self-hosted, Go-based SFU)** + coturn | Group calls, meetings, screen share, recording (egress). Recommended over building an SFU from scratch |
| Database | **Neon (Postgres 16+)** | RLS for tenancy; Neon branching for per-PR preview environments; logical replication out for analytics |
| Cache/queues | Redis (presence, sessions, rate limits) + **NATS JetStream** (events, workflow engine, webhooks) | |
| Search | Meilisearch (records + messages) | |
| Object storage | S3-compatible (Cloudflare R2 / AWS S3) | Attachments, recordings, exports |
| Email | SMTP out via provider (SES/Postmark); inbound parsing for email-to-case | |
| Deploy | Docker + Kubernetes (or Nomad) on Hetzner/AWS; Cloudflare CDN/WAF | Media servers need UDP + public IPs — plan networking early |
| CI/CD | GitHub Actions; migrations via Atlas/golang-migrate; Neon branch-per-PR | |

### 7.2 Architecture decisions & rationale
1. **Modular monolith over microservices at start.** One Go binary, strict module boundaries (`/modules/sales`, `/modules/finance`, …) with internal interfaces and a shared event bus. Extract chat/media/workflow into services only when load requires. 10+ devs can still parallelize via module ownership.
2. **Metadata-driven entity layer.** A `entity_definitions`/`field_definitions` catalog drives dynamic tables (one physical table per entity, migrated on publish — not EAV) so custom entities stay fast and indexable.
3. **Tenancy = org_id column + Postgres RLS everywhere**, with BU-scope checks in the authorization service (privilege-depth model evaluated in Go, cached in Redis). Never rely on RLS alone or app checks alone — both.
4. **Buy/adopt the media layer.** LiveKit gives SFU, simulcast, recording egress, and SDKs (client JS + Go server SDK). Building an SFU in-house is a 12-month detour with worse results. Signaling/auth/rooms remain our code; media transport is LiveKit.
5. **Neon specifics:** connection pooling via Neon's pooler (Fiber services use pgx pool); watch for cold-start latency on autosuspend — keep production compute always-on; use read replicas for dashboards/reporting; branching powers ephemeral test environments.
6. **Event-driven side effects.** Every write emits a domain event to NATS; workflow engine, audit log, search indexer, notifications, and webhooks are all consumers. Keeps the write path fast and modules decoupled.
7. **API-first.** The Nuxt app consumes the same versioned REST API exposed to customers. No private endpoints.

### 7.4 Module entitlement architecture
Each Go module package exports a `Manifest` (entities, routes, nav, widgets, privileges, event subscriptions) registered at boot into a module registry. Tenant entitlements live in `identity.org_subscriptions`; the authorization layer computes an **effective manifest** per request context (tenant ∩ BU ∩ role), cached in Redis and invalidated on entitlement change. HTTP middleware resolves the route's owning module and rejects disabled ones (404, not 403 — disabled modules are invisible). The frontend fetches `GET /me/manifest` at session start; the Nuxt shell builds navigation, dashboard widget registry, and search scope from that payload, with shadcn-vue components rendering whatever the manifest declares. Result: adding a module to a tenant is a data change, visible on next refresh.

### 7.5 Push notification pipeline
Domain events (NATS) → notification service evaluates per-user rules/preferences → fan-out: WebSocket (app open), Web Push via stored VAPID subscriptions (app closed), email digest (fallback). Device subscriptions stored per user with liveness pruning on push failure. Incoming-call pushes are high-priority with short TTL (a stale ring is worse than none).

### 7.3 High-level data domains
`identity` (users, orgs, BUs, teams, roles) · `metadata` (entities, fields, forms, views) · `crm` (accounts, contacts, leads, opportunities, products, quotes, orders) · `service` (cases, queues, SLAs, entitlements, KB) · `finance` (accounts, journals, invoices, bills, payments, tax) · `scm` (items, warehouses, stock, POs, shipments) · `projects` (projects, tasks, timesheets) · `hr` (employees, leave) · `collab` (conversations, messages, calls, meetings, presence) · `automation` (workflows, runs) · `audit`

## 8. Phased Roadmap & Scope

### Phase 0 — Foundations (Months 1–3)
Platform skeleton: auth + 2FA, org/BU/team/role model, RBAC engine, metadata entity layer, **module registry + tenant entitlements**, base UI shell (shadcn-vue design system, entitlement-driven nav, grids, forms), audit, files, API framework, CI/CD, environments. **Exit:** create a custom entity with fields, secure it with roles, CRUD it through generated UI and API — and toggle a module off for a tenant and watch it vanish from nav and API.

### Phase 1 — CRM + Service + Chat & 1:1 Calls (Months 4–9) — MVP
Sales module (5.2), Customer Service core (cases, queues, SLAs, email-to-case, live chat widget, KB), collaboration chat + presence + 1:1 audio/video, dashboards, workflow engine v1, **PWA install + web push notifications**, global search. **Exit criterion: Andrew's BPO runs a pilot client entirely on the platform — no Dynamics, no Teams for internal chat — with agents on the installed PWA receiving push for assignments and calls.**

### Phase 2 — Meetings, Group Calls, Projects, HR-lite (Months 10–14)
Meeting links + guest browser join + lobby, scheduling/ICS, group calls (SFU), screen share, supervisor console v2 (barge/whisper on chat), skills-based routing, customer portal, Project Ops, HR-lite, CSAT.

### Phase 3 — Finance + Supply Chain + Recording + Voice channel (Months 15–22)
GL/AR/AP/tax/banking/expenses/budgets/financial reports, procurement + inventory + fulfilment, meeting recording, WebRTC voice channel into service queues, entitlements billing, utilization → invoicing pipeline.

### Phase 4 — Hardening & Expansion (Months 23+)
PSTN/SIP trunks, marketing-lite, SSO (SAML/OIDC), advanced reporting/BI export, mobile apps, marketplace/API ecosystem, SOC 2 readiness.

### Out of scope for the entire roadmap (restated)
Manufacturing, field service, payroll engine, AI copilots (until post-Phase 4), on-prem packaging, feature-parity commitments with any specific Dynamics SKU.

## 9. Team & Developer Breakdown (12–14 people)

### 9.1 Squad structure

| Squad | Headcount | Ownership |
|---|---|---|
| **Platform Core** | 3 backend (Go) | Identity, RBAC, tenancy/RLS, metadata engine, workflow engine, audit, API framework, search, notifications |
| **Business Apps** | 2 backend (Go) + 2 frontend (Nuxt) | Sales, Customer Service (then Finance, SCM, Projects, HR in later phases) |
| **Collaboration/RTC** | 2 backend (Go, WebSocket + LiveKit integration) + 1 frontend (Nuxt, WebRTC client) | Chat, presence, calls, meetings, media infra |
| **Frontend Platform** | 1 senior frontend | App shell, design system, form/grid/dashboard engines shared by all modules |
| **DevOps/SRE** | 1 | K8s, Neon management, LiveKit/coturn infra, CI/CD, observability, backups, security ops |
| **QA/Test Eng** | 1 | E2E framework (Playwright), tenancy-isolation test suite, load tests (incl. call load) |
| **Product/Design** | 1 PM (can be Andrew) + 1 designer | Requirements, Dynamics feature mapping, UX |

Tech leadership: one of the Platform Core seniors doubles as architect. In later phases, Business Apps splits into a Finance/SCM squad and a CRM/Service squad by hiring +2–3.

### 9.2 Epic-level work breakdown (Phases 0–1)

| # | Epic | Squad | Est. (dev-weeks) |
|---|---|---|---|
| E1 | Auth, sessions, 2FA, org/user management | Platform | 8 |
| E2 | BU hierarchy, teams, security roles, privilege-depth authz engine + Redis cache | Platform | 10 |
| E3 | Metadata entity/field catalog + dynamic migrations + generated CRUD API | Platform | 12 |
| E4 | RLS + tenancy enforcement + isolation test harness | Platform + QA | 6 |
| E4b | Module registry, tenant entitlements, effective-manifest service + admin console | Platform | 8 |
| E5 | App shell, entitlement-driven nav, shadcn-vue design system, per-tenant theming | FE Platform | 8 |
| E6 | Grid/view engine (saved queries, sorting, filtering, export) | FE Platform | 8 |
| E7 | Form engine (JSON-schema forms, layout designer v1) | FE Platform + Biz FE | 10 |
| E8 | Audit log + activity timeline + files/attachments | Platform | 6 |
| E9 | Workflow engine v1 (triggers, conditions, actions, run log) | Platform | 10 |
| E10 | Notification pipeline (in-app, email, web push fan-out, preferences) | Platform | 6 |
| E10b | PWA: manifest, service worker, offline shell, install flows, badging, update prompt | FE Platform | 5 |
| E11 | Global search (Meilisearch indexing pipeline) | Platform | 4 |
| E12 | Sales: leads/accounts/contacts/opportunities + pipeline UI | Biz Apps | 12 |
| E13 | Sales: products, price lists, quotes/orders, PDF generation | Biz Apps | 8 |
| E14 | Email integration (send/receive, email-to-case, templates) | Biz Apps | 8 |
| E15 | Service: cases, queues, routing, SLA engine | Biz Apps | 12 |
| E16 | Service: KB, entitlements, CSAT | Biz Apps | 6 |
| E17 | Live chat widget (embeddable) + agent inbox | Collab + Biz Apps | 8 |
| E18 | Chat core: conversations, channels, threads, files, search, retention | Collab | 12 |
| E19 | WebSocket gateway, presence, offline queue, push | Collab | 8 |
| E20 | 1:1 calls: signaling, P2P WebRTC, TURN, call UI, call logging | Collab | 10 |
| E21 | Dashboards/charts engine | FE Platform + Biz Apps | 6 |
| E22 | Infra: K8s, environments, CI/CD, Neon branching, observability | DevOps | 10 |
| E23 | E2E + load test framework; security review | QA | 8 |

≈ 209 dev-weeks ≈ 9–10 months with ~10 delivery engineers at 60 % focus factor — consistent with the Phase 0–1 timeline.

### 9.3 Phase 2–3 epic highlights
Meetings & guest join (10 wk) · LiveKit SFU group calls + screen share (10 wk) · Scheduling/calendar/ICS (6 wk) · Supervisor console v2 (6 wk) · Projects + timesheets (10 wk) · HR-lite (6 wk) · GL/journals/CoA (10 wk) · AR/AP + payments (12 wk) · Tax + multi-currency (8 wk) · Banking + reconciliation (6 wk) · Inventory + procurement (14 wk) · Fulfilment (8 wk) · Recording/egress (4 wk) · Voice channel + queue integration (8 wk).

### 9.4 Engineering practices
Trunk-based development; PR review required; Neon branch per PR with seeded data; contract tests on the public API; tenancy-isolation suite runs on every merge (non-negotiable); load test call infrastructure quarterly; ADRs for architectural decisions; feature flags for progressive rollout per tenant.

## 10. Success Metrics

| Metric | Target |
|---|---|
| BPO pilot client fully operated on platform | End of Phase 1 |
| Agent handle time vs. Dynamics baseline | ≤ parity within 60 days of migration |
| Chat message delivery P95 | < 500 ms |
| Call setup success rate | > 98 % |
| Monthly infra cost per seat at 50 seats | < 40 % of equivalent Dynamics + Teams licensing |
| SLA timer accuracy | 100 % (zero missed breach actions) |
| Cross-tenant data leakage incidents | 0, ever |

## 11. Risks & Mitigations

| Risk | Likelihood | Mitigation |
|---|---|---|
| **Scope gravity** — "all Dynamics features" is unachievable; team drowns | High | This PRD's phase gates are contractual; new features enter a backlog reviewed per phase, never mid-phase. Ship the BPO-critical path first. |
| WebRTC/media complexity underestimated | High | Adopt LiveKit + coturn rather than building media; hire at least one engineer with prior WebRTC production experience |
| Metadata engine becomes a performance/complexity trap | Medium | Real tables per entity (no EAV); publish-time migrations; load-test with 10 M rows early |
| Neon autosuspend/pooling surprises under WebSocket-heavy load | Medium | Always-on compute for prod; pgx pooling tuned; chat/presence state in Redis, not Postgres |
| Finance correctness (double-entry, tax, FX) | Medium | Hire/contract an accountant-domain expert for Phase 3; property-based tests on ledger invariants; immutable journals |
| Key-person dependency on architect | Medium | ADRs, docs, pairing rotation |
| Building this while also running the BPO on Dynamics | High | Run Dynamics in parallel through Phase 1; migrate one pilot client only; keep the Dynamics exit optional until parity is proven |
| Compliance (GDPR/Ghana DPA) gaps discovered late | Low | DPIA at Phase 0; data-map maintained per module |

## 12. Open Questions
1. Commercial model: internal tool only, or SaaS productization from Phase 2? (Affects billing/metering priority.)
2. Hosting region strategy: single EU region vs. EU + Africa (Neon region availability for Ghana-latency users; media servers can be regional sooner).
3. Migration tooling from Dynamics 365 (accounts/contacts/cases export → import maps) — Phase 1 requirement or manual one-off?
4. Whether the customer chat widget and the internal chat share one message store or are separated for retention/compliance reasons.
5. Build vs. adopt for calendar/scheduling (e.g., embed an open-source scheduling lib vs. in-house).

---
*Prepared for Andrew — Ingen Cloud Technologies. Draft v1.0 for review.*
