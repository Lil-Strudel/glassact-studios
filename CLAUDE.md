# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project

GlassAct Studios is a B2B platform for ordering custom stained glass inlays for gravestones. Dealerships (engravers) submit projects, GlassAct internal users design proofs, manufacture inlays, and invoice. Detailed domain rules (project/proof/manufacturing status flows, ordering, invoicing, notifications) live in `.cursor/rules/domain.md` — read it before touching business logic.

## Repository Layout

pnpm workspaces + Go modules monorepo.

- `apps/api` — Go HTTP API (`net/http` + `http.ServeMux` + `justinas/alice` middleware). Entry point `cmd/api/main.go`. Feature modules under `modules/<name>/` each expose `*Module` with handlers/services; routes registered in `modules/modules.go`.
- `apps/webapp` — SolidJS SPA (Vite, TanStack Router/Query/Form/Table, Tailwind, Kobalte). Authenticated user app.
- `apps/landing` — Astro marketing site.
- `apps/infrastructure` — Terraform (AWS Lambda via `aws-lambda-go-api-proxy`, S3, certs).
- `libs/data` — **Shared data layer.** TS types (`src/`) consumed by frontend; Go models (`pkg/`) consumed by API. SQL migrations (`migrations/`) and Jet-generated code (`pkg/gen/`, do not edit).
- `libs/ui` — SolidJS component library (`@glassact/ui`).

The Go module is rooted at the repo, not `apps/api` — `go test ./...` from repo root sees everything.

## Common Commands

Run from repo root unless noted.

**Dev stack (preferred):** `pnpm dev` — runs mprocs, which starts Postgres (docker, postgis/postgis:18-3.6), mailpit, api (`air` live reload), and webapp. Landing has `autostart: false`. Postgres listens on 5432 (user `dev`, pass `1234`, db `glassact`); mailpit SMTP on 1025, UI on 8025.

**Individual apps:** `pnpm dev:api`, `pnpm dev:webapp`, `pnpm dev:landing`.

**Database:**
- `pnpm db:migrate/new <name>` — scaffold a new migration pair
- `pnpm db:migrate/up` / `db:migrate/down` / `db:migrate/goto` / `db:migrate/force`
- `pnpm db:gen` — regenerate Jet code from live DB schema (run after migrations)
- `pnpm db:seed` — load `seed.sql`
- `pnpm db:psql` — psql shell
- Each data command sources `libs/data/.env` for `DATABASE_DSN`.

**Tests:**
- `pnpm api:test` — API handler tests (`go test ./modules/...` inside `apps/api`)
- `pnpm db:test` — data-layer Go tests (`go test ./libs/data/pkg/...` from repo root) — use testcontainers, so Docker must be running
- `pnpm db:test:race` / `db:test:short` / `db:test:bench` for variants
- Single test: `go test -v -run TestName ./path/to/pkg`

**Lint:** `pnpm lint` (all TS packages) or targeted `pnpm lint:webapp` / `lint:ui`.

**Build libs:** `pnpm libs:build` before consuming built `dist/` output of `@glassact/data` / `@glassact/ui`.

## Architecture Notes

### StandardTable pattern
Every primary entity carries `id` (int), `uuid` (string), `created_at`, `updated_at`, `version` — enforced both in TS (`StandardTable<T>` in `libs/data/src/helpers.ts`) and Go (`data.StandardTable` in `libs/data/pkg/helpers.go`). `version` is used for optimistic locking on UPDATE — always include it in WHERE and in the SET clause.

### Dual persistence handles
Go models hold both `DB *pgxpool.Pool` and `STDB *sql.DB`. Jet queries use `STDB`; raw pgx calls use the pool. Transactional variants are `TxInsert`/`TxUpdate`/... taking `*sql.Tx` — use these inside `m.Db.STDB.Begin()` blocks.

### API model ↔ Jet model conversion
For each entity: `xxxFromGen(model.Xxx) *Xxx` and `xxxToGen(*Xxx) (*model.Xxx, error)`. Handlers and the rest of the app only see the API struct; the generated `model.*` types never leak out of `libs/data/pkg/`. Nullable pointer fields require the explicit nil-check/copy dance seen in `projects.go`.

### Enum pattern
String-literal unions in TS (with a `PROJECT_STATUSES` array for runtime) mirror a Go `type XxxStatus string` plus a `XxxStatuses` struct of consts. SQL enforces via `CHECK` constraints, not Postgres enums. Keep the three in sync.

### Dual auth system
Two user tables: `dealership_users` (B2B customers, multi-tenant by `dealership_id`) and `internal_users` (GlassAct staff). Both implement the `data.AuthUser` interface (`GetID`, `IsDealership`, `IsInternal`, `Can(action)`, ...). `app.Authenticate` middleware checks dealership table first, then internal, and stashes the user on request context. Handlers fetch via `m.ContextGetUser(r)` (generic) or the typed helpers (`ContextGetDealershipUser`/`ContextGetInternalUser`, which panic on wrong type). Permission-gated routes use `app.RequirePermission(data.ActionXxx)` in the alice chain. OAuth callbacks are invite-only — a user must already exist in one of the tables.

### Multi-tenancy
Dealership-user requests must be scoped to their `dealership_id` in every query. Never trust a client-supplied `dealership_id` for a dealership user — compare against `user.DealershipID`. Internal users bypass this scope.

### Frontend query layer
Queries live in `apps/webapp/src/queries/`. Each entity file exports raw fetchers (`getXxx`), `queryOptions` factories (`getXxxOpts`), and `mutationOptions` factories. Types come from `@glassact/data` using `GET<T>` / `POST<T>` / `PATCH<T>` helpers — never redefine API shapes in the webapp. Query key convention: `[entity]`, `[entity, uuid]`, `[entity, uuid, nested]`, `[entity, {filter}]`.

### SolidJS reactivity (not React!)
- Signals are functions — call `count()`, never reference bare.
- Never destructure props at function scope (breaks reactivity). Access as `props.x`, wrap derived values in `createMemo`, or compose with `splitProps`.
- Prefer `createMemo` over `createEffect` for derived state.
- Permission UI gates use `<Can permission="...">`, not inline role checks.

### Data-sync checklist (adding an entity)
1. SQL migration (table + triggers for `updated_at`/`version`)
2. `pnpm db:migrate/up` then `pnpm db:gen`
3. TS type in `libs/data/src/<entity>.ts`, exported from `index.ts`
4. Go model in `libs/data/pkg/<entity>.go` with the API struct, `FromGen`/`ToGen`, CRUD, and Tx variants
5. Register model in `libs/data/pkg/models.go`
6. Feature module in `apps/api/modules/<entity>/` + route wiring in `modules.go`

## Scoped CLAUDE.md Files

Deeper conventions live alongside the code they govern:

- `apps/api/CLAUDE.md` — Go backend: module pattern, handler/data-layer templates, auth middleware, validator/v10, testcontainers.
- `apps/webapp/CLAUDE.md` — SolidJS frontend: reactivity rules, query/mutation layout, `<Can>` permission component, TanStack Form, route conventions.
- `libs/data/CLAUDE.md` — Shared data layer: StandardTable pattern, TS type helpers (`GET`/`POST`/`PATCH`), Go model template, Jet regeneration.
- `libs/ui/CLAUDE.md` — UI library specifics (short — inherits from the webapp conventions).
- Domain rules (project/proof/manufacturing flows, ordering, invoicing, notifications) are captured below.

## General Conventions

### Philosophy
- Correctness over cleverness. If a solution requires explanation, consider simplifying it.
- Dependencies must justify their weight. Check the standard library first; for simple utilities, write them. Avoid micro-packages (`is-odd`, `left-pad`) and packages wrapping stdlib with minimal value.
- **Approved high-value dependencies:** TanStack libraries (Query, Router, Form, Table), Kobalte, Zod, Jet (SQL building), pgx (Postgres driver), validator/v10.

### Code style
- No redundant comments. Comments explain WHY, never WHAT — prefer more descriptive identifiers over a comment. Only comment things a developer could not infer from reading the code.
- Be descriptive over brief. Booleans read as questions (`isActive`, `hasPendingProof`). Functions describe actions (`createProof`, `declineProof`).
- One primary export per file. File size soft limit: ~300 lines before splitting.
- Handle errors explicitly — never swallow. Wrap with context: `fmt.Errorf("failed to create proof for inlay %d: %w", inlayID, err)`.

### Testing
- Test behavior, not implementation. Tests coupled to internals break on harmless refactors.
- Prefer integration tests with testcontainers where practical; unit-test complex business logic.
- Name Go tests descriptively: `TestCreateProof_WithMissingInlay_ReturnsError`.

### Git
- Committing is a human job. Suggest commit messages; do not run `git commit` unless the user explicitly asks.
- Commit subject: imperative mood, ≤72 chars. Body explains WHY.
- Branch naming: `feature/…`, `fix/…`, `refactor/…`.

### Security
- Never commit secrets. Ensure `.env.example` covers required vars.
- Validate external input at API boundaries. Use strong typing to prevent invalid states.
- **Multi-tenancy:** every dealership-scoped query must filter by `dealership_id`. Never trust a client-provided `dealership_id` for a dealership user — compare against `user.DealershipID`. Test permission boundaries explicitly.

---

# Domain Rules

Business logic and domain constraints for the GlassAct Studios platform. Read this before touching project/proof/manufacturing/invoice logic.

## Business Overview

GlassAct Studios manufactures custom stained glass inlays for gravestones. The platform serves B2B customers (gravestone engravers called "dealerships") who order inlays on behalf of end consumers.

### Key Stakeholders

| Stakeholder         | Role                                           |
| ------------------- | ---------------------------------------------- |
| Dealership          | Orders inlays, approves designs, pays invoices |
| GlassAct Designer   | Creates proofs, responds to design feedback    |
| GlassAct Production | Manages manufacturing workflow                 |
| GlassAct Billing    | Creates and manages invoices                   |

## Entity Lifecycles

### Project Status Flow

```
┌───────┐    ┌─────────┐    ┌───────────────┐    ┌─────────┐    ┌──────────┐    ┌───────────┐
│ draft │───►│ ordered │───►│ in-production │───►│ shipped │───►│ invoiced │───►│ completed │
└───────┘    └─────────┘    └───────────────┘    └─────────┘    └──────────┘    └───────────┘

draft │ ordered ───► cancelled   (in-production onwards cannot be cancelled)
```

The project itself only has two pre-order states: `draft` (project exists, building inlays) and `ordered` (order placed, manufacturing about to start). Inlay-level readiness — not a project-level status — gates the Place Order button.

| Status        | Description                            | Actions Available                                                                  |
| ------------- | -------------------------------------- | ---------------------------------------------------------------------------------- |
| draft         | Project is being built / inlays added  | Add/remove inlays, customize, edit name/internal_reference, cancel, place order    |
| ordered       | Order placed, queued for production    | Cancel (only here and in `draft`)                                                  |
| in-production | Manufacturing in progress              | Track milestones                                                                   |
| shipped       | All inlays shipped                     | Mark delivered — moves to `invoiced`, or straight to `completed` if already paid   |
| invoiced      | Invoice sent                           | Pay                                                                                |
| completed     | Payment received                       | -                                                                                  |
| cancelled     | Project cancelled                      | - (terminal)                                                                       |

### Inlay Readiness

Place Order is enabled only when **every** inlay on the project is "ready". An inlay is ready if any of:

- **Stock catalog inlay** (`type='catalog'`, `is_customized=false`) — ready immediately. Priced at the catalog item's `default_price_group_id`.
- **Customized catalog inlay** (`type='catalog'`, `is_customized=true`) — produced via the customizer. Becomes ready once the auto-created `inlay_proof` (with `approval_authority='internal'`) is approved by an internal designer/admin, who may also override the price group.
- **Custom inlay** (`type='custom'`) — becomes ready once an internal designer creates a proof (with `approval_authority='dealership'`) and the dealership approves it. This is the only flow that still uses the legacy "designer uploads → customer reviews" loop.

### Project Fields

- `name` (required) — display name.
- `internal_reference` (optional, nullable TEXT) — dealership's PO number / internal job reference. Surface alongside the project name everywhere.
- `installation_kit` (BOOLEAN) — whether the order includes an installation kit. **One kit per project, not per inlay**: it covers the install materials for every inlay on the project and is a single flat `INSTALLATION_KIT_PRICE_CENTS` charge regardless of inlay count. Editable only while the project is `draft`.
- `installation_kit_price_cents` (nullable INTEGER) — NULL until the order is placed, then locked to the charge, the project-level analogue of `order_snapshots.price_cents`.

### Chat

**One chat thread per project** (`project_chats`), shared by every inlay on it. There is no per-inlay thread.

- `inlay_id` (nullable FK) **tags** a message with the inlay it is about — it is not a scope. Untagged messages are about the project as a whole.
- Every tagged message renders the inlay's name as a link to that inlay, on the project page and the inlay page alike, so "this one is too dark" is never ambiguous.
- The project page shows the thread in a right-hand rail; the inlay page shows the same thread filtered to that inlay, with a toggle back to the whole project.
- Proof events (`proof_sent` / `proof_approved` / `proof_declined`) land in this thread tagged with their inlay.
- Deleting an inlay leaves its messages in the thread, untagged (`ON DELETE SET NULL`).

### Proof Status Flow

```
┌─────────┐     ┌──────────┐
│ pending │────►│ approved │ (terminal)
└────┬────┘     └──────────┘
     │
     │          ┌──────────┐
     └─────────►│ declined │───► (new proof created)
                └──────────┘

     │          ┌────────────┐
     └─────────►│ superseded │ (when newer version exists)
                └────────────┘
```

- A proof starts as `pending` when created.
- `approved` is terminal — cannot be changed.
- `declined` triggers feedback; designer creates new proof.
- When a new proof is created, previous `pending` proofs become `superseded`.

**Approval authority** (`inlay_proofs.approval_authority`):
- `dealership` — customer-facing proof (custom inlays). An approver/admin dealership user approves or declines; the approval lands in the project chat thread (tagged with the inlay) and notifies internal staff.
- `internal` — customizer-baked proof (customized catalog inlays). An internal designer/admin (with `internal_approve_proof` permission) approves and may override `price_group_id`. No chat message; no customer involvement.

### Manufacturing Steps

```
ordered → materials-prep → cutting → fire-polish → packaging → ready-to-ship
```

- The ladder stops at `ready-to-ship`. Shipping and delivery are recorded on the **project**, not the inlay — there is no `shipped` or `delivered` manufacturing step.
- Steps can move backward (via "revert" milestone events).
- Each transition creates an `inlay_milestone` record.
- Progress is event-based, not a single status field.
- Step transitions are unconditional — nothing blocks an inlay from moving. Internal users post `inlay_updates` to communicate issues (see Inlay Updates below).

## Business Rules

### Ordering

**Order placement requires:** project is in `draft`; user has `place_order` permission (dealership approver or admin); every selected inlay is ready (see Inlay Readiness above). The dealership user picks which inlays to include in the cart — only the selected ones get an `order_snapshot` and enter manufacturing.

**Price locking.** When an order is placed, one `order_snapshot` is created per selected inlay:

| Inlay flavor                           | `proof_id`           | `price_group_id` / `price_cents` source                                          | `width` / `height` source       |
| -------------------------------------- | -------------------- | -------------------------------------------------------------------------------- | ------------------------------- |
| Stock catalog                          | NULL                 | catalog item's `default_price_group_id` and that group's `base_price_cents`      | catalog item defaults           |
| Customized catalog (approved internal) | approved proof's ID  | from the approved proof (price_group_id; price_cents if set else group base)     | from the approved proof         |
| Custom (approved dealership)           | approved proof's ID  | same                                                                             | from the approved proof         |

Snapshot values are immutable. Invoices read from snapshots, not from current catalog or proof state.

The installation kit is **not** on the snapshot — it is one flat project-level charge, locked into `projects.installation_kit_price_cents` in the same transaction. A project total is therefore `SUM(order_snapshots.price_cents) + COALESCE(projects.installation_kit_price_cents, 0)`.

**Prices exclude tax and shipping.** Nothing in the schema prices either yet, so every amount shown to a dealership in a project context carries the `PRICE_CAVEAT` line ("Before tax and shipping.") — once per surface, under the total.

### Proofs

- **Price group is assigned at the proof level**, not the inlay level. A catalog item has `default_price_group_id`; the designer may override based on custom sizing, customization complexity, or special materials.
- **Versioning:** proofs are versioned per inlay — `(inlay_id, version_number)` is unique. All versions are visible to the dealership.
- **Proof-chat integration:** when a proof is created, also insert a `project_chats` message with `message_type = 'proof_sent'`, `inlay_id` set to the inlay, and link `proof.sent_in_chat_id = chat_message.id`. Supersede previous `pending` proofs on the same inlay. Update `inlay.preview_url`. Notify dealership.

### Manufacturing

**Milestone events** (progress is event-based):

| Event Type | Meaning                           |
| ---------- | --------------------------------- |
| entered    | Inlay arrived at this step        |
| exited     | Inlay moved to next step          |
| reverted   | Inlay moved backward to this step |

Example:
```
1. entered:ordered          (order placed)
2. exited:ordered           (starting materials)
3. entered:materials-prep
4. exited:materials-prep
5. entered:cutting
6. reverted:materials-prep  (problem found, going back)
7. exited:materials-prep
8. entered:cutting
```

`inlays.manufacturing_step` is stored for query convenience; the milestone history is the source of truth.

**Inlay Updates:**

Internal users (production/admin, `create_inlay_update` permission) post `inlay_updates` to keep the dealership informed about what's happening to an inlay. Updates are informational only — they **never** block a step transition and have no resolution lifecycle.

| Type  | Meaning                                              |
| ----- | ---------------------------------------------------- |
| info  | General note (e.g. "ahead of schedule")              |
| issue | Something went wrong / needs rework (e.g. "dropped during fire-polish, restarting from materials-prep") |

Each update carries a free-text `message` and a `step` (the inlay's manufacturing step when posted, for context). Updates are shown interleaved chronologically with milestone events on the inlay's manufacturing timeline, and posting one notifies the dealership (`inlay_update`).

### Users & Permissions

**Multi-tenancy.** Dealership users see only their own dealership's data — scope every query by `dealership_id`. Internal users see all dealerships' data.

**Dealership user roles:**

| Role      | Can Do                                 |
| --------- | -------------------------------------- |
| viewer    | View projects, chats, invoices         |
| submitter | + Create projects, add inlays, chat    |
| approver  | + Approve/decline proofs, place orders |
| admin     | + Manage users, pay invoices           |

**Internal user roles:**

| Role       | Can Do                                 |
| ---------- | -------------------------------------- |
| designer   | Create proofs, respond to design chats |
| production | Manage kanban, post inlay updates       |
| billing    | Create invoices, mark paid             |
| admin      | Everything                             |

### Invoicing

- Invoices are 1:1 with projects.
- Cannot create invoice until project is ordered.
- Line items auto-populated from order snapshots; additional line items (shipping, fees) can be added manually.
- Full payment only (no partial payments in MVP).
- Invoice uses snapshot prices, not current catalog prices.

### Notifications

The full set is `NotificationEventTypes` in `libs/data/pkg/notifications.go`.

**Recipients are watch-list driven** (GitHub's model for issues/PRs). Every send goes through `NotifyDealership` or `NotifyInternal` in `apps/api/app/notifications.go`, each of which resolves recipients as:

```
watchers on that side of the project  ∪  role fallback for this event type  −  the actor
```

The **actor is never notified of their own action**, in-app or by email.

`project_watchers` (one row per user per project) holds the subscription. No row = never subscribed; `is_watching = true` = watching; `is_watching = false` = explicitly unwatched, which auto-subscription must never undo (`AutoSubscribe` uses `ON CONFLICT DO NOTHING`).

**Auto-subscribe** — `app.AutoWatchProject(projectID, user)` runs whenever someone does something meaningful: creating a project, placing an order, posting a chat message, creating/approving/declining a proof, submitting a custom or customized inlay, advancing a manufacturing step, posting an inlay update, shipping/delivering, and creating/voiding/paying an invoice.

**Role fallback** — a small set of events must reach whoever is allowed to act on them even if nobody is watching, so work never sits unclaimed. Defined as `internalRoleFallback` / `dealershipRoleFallback` in `apps/api/app/notifications.go`.

| Event                    | Recipients                                | Description                                     |
| ------------------------ | ----------------------------------------- | ----------------------------------------------- |
| proof_ready              | Watchers + dealership approver, admin      | New proof available                             |
| proof_approved           | Internal watchers                          | Proof was approved                              |
| proof_declined           | Internal watchers                          | Proof was declined                              |
| internal_review_required | Watchers + internal designer, admin        | A customizer-baked proof needs internal pricing |
| custom_inlay_submitted   | Watchers + internal designer, admin        | A custom inlay needs a proof                    |
| order_placed             | Watchers + internal production, admin      | New order in queue                              |
| inlay_step_changed       | Dealership watchers                        | Inlay moved in manufacturing                    |
| inlay_update             | Dealership watchers                        | New update posted on inlay                      |
| project_shipped          | Dealership watchers                        | Project shipped                                 |
| project_delivered        | Dealership watchers; watchers + internal billing, admin | Delivery confirmed; project moved to invoiced |
| invoice_sent             | Watchers + dealership admin                | Invoice available                               |
| invoice_voided           | Watchers + dealership admin                | Invoice was voided                              |
| payment_received         | Watchers + dealership admin                | Payment confirmed                               |
| chat_message             | Watchers on the other side of the chat     | New message                                     |

Users can disable specific notification types; disabled notifications still appear in-app, just no email is sent. That preference is orthogonal to watching — the watch list decides *whether* you are a recipient, the preference decides whether that recipient also gets an email.

## Catalog

- Catalog items have unique `catalog_code` (e.g. "A-BRD-0003L"), default and minimum dimensions, a default price group, tags, and multiple images (one primary).

| Aspect             | Catalog Inlay                 | Custom Inlay                    |
| ------------------ | ----------------------------- | ------------------------------- |
| Reference          | `catalog_item_id`             | description + reference images  |
| Initial dimensions | From catalog defaults         | Customer's requested dimensions |
| Customization      | `customization_notes`         | Full custom design              |
| Pricing basis      | Catalog default + adjustments | Designer assessment             |

## Future Considerations

**Graphical editor (post-MVP).** `inlay_proofs.scale_factor` and `inlay_proofs.color_overrides` are pre-wired; the editor will start from the catalog item's design asset, apply scale + color overrides, and regenerate `preview_url`.

**Per-inlay pricing (post-MVP).** `inlay_proofs.price_cents` is nullable now (price derived from `price_group_id`). When set, order snapshot captures whichever is present.
