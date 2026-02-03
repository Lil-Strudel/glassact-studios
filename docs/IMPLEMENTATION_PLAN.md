# GlassAct Studios - Implementation Plan

This document outlines the complete implementation plan for the GlassAct Studios ecommerce platform. It covers data models, feature breakdown, task dependencies, and development phases.

## Table of Contents

1. [Executive Summary](#executive-summary)
2. [Architecture Overview](#architecture-overview)
3. [Data Models](#data-models)
4. [Feature Breakdown](#feature-breakdown)
5. [Implementation Phases](#implementation-phases)
6. [Task Dependencies](#task-dependencies)
7. [MVP vs Post-MVP](#mvp-vs-post-mvp)

---

## Executive Summary

GlassAct Studios is a B2B ecommerce platform for custom stained glass inlays in the memorial industry. The platform serves gravestone engravers ("dealerships") who order inlays on behalf of their customers.

### Key Differentiators

- **Conversational ordering**: Each inlay has its own design discussion thread
- **Non-linear manufacturing**: Inlays can move backward in the production process
- **Dual user systems**: Separate models for dealership users and internal staff
- **Proof-centric workflow**: Designs are versioned with explicit approval tracking

### Core Decisions

| Decision | Choice | Rationale |
|----------|--------|-----------|
| User Model | Separate tables for dealership/internal | Clean separation, different permission models |
| Permissions | Role-based presets | Simpler MVP, centralized for future expansion |
| Manufacturing | Event-based milestones | Supports non-linear progression |
| Proofs | Versioned, in-chat | Natural conversation flow |
| Pricing | Locked at order | Immutable snapshots for billing accuracy |
| Notifications | Email only, polling | Simpler MVP, websockets later |

---

## Architecture Overview

### System Boundaries

```
┌─────────────────────────────────────────────────────────────────┐
│                        DEALERSHIP SIDE                          │
├─────────────────────────────────────────────────────────────────┤
│  Dealership Users                                               │
│  ├── Browse catalog                                             │
│  ├── Create projects & add inlays                               │
│  ├── Chat about designs                                         │
│  ├── Approve/decline proofs                                     │
│  ├── Place orders                                               │
│  ├── Track manufacturing progress                               │
│  └── View/pay invoices                                          │
└─────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────┐
│                        GLASSACT SIDE                            │
├─────────────────────────────────────────────────────────────────┤
│  Internal Users                                                 │
│  ├── Manage catalog                                             │
│  ├── Create proofs for inlays                                   │
│  ├── Respond to design chats                                    │
│  ├── Manage manufacturing kanban                                │
│  ├── Create/resolve blockers                                    │
│  ├── Create and send invoices                                   │
│  └── View dashboards and reports                                │
└─────────────────────────────────────────────────────────────────┘
```

### Tech Stack

| Layer | Technology |
|-------|------------|
| Frontend | SolidJS, TanStack (Router, Query, Form, Table), Tailwind |
| Backend | Go, standard library HTTP, Jet SQL builder |
| Database | PostgreSQL with PostGIS |
| Storage | AWS S3 for images and design assets |
| Monorepo | pnpm workspaces |

### Package Structure

```
glassact-studios/
├── apps/
│   ├── api/           # Go backend
│   ├── webapp/        # SolidJS dealership + internal app
│   └── landing/       # Marketing site (Astro)
├── libs/
│   ├── data/          # Shared types (TS) + data layer (Go)
│   └── ui/            # SolidJS component library
└── docs/              # Documentation
```

---

## Data Models

### Entity Relationship Diagram

```
DEALERSHIPS
    └── dealership_users ──────────────────┐
    └── projects                           │
            └── inlays                     │
            │       └── inlay_chats ◄──────┤ (sender)
            │       └── inlay_proofs       │
            │       └── inlay_milestones   │
            │       └── inlay_blockers     │
            └── project_chats ◄────────────┤ (sender)
            └── invoices                   │
            └── order_snapshots            │
                                           │
INTERNAL_USERS ────────────────────────────┘
    └── (performs manufacturing actions)
    └── (sends chat messages)
    └── (creates proofs)

CATALOG_ITEMS
    └── catalog_item_tags
    └── catalog_item_images

PRICE_GROUPS
    └── (referenced by catalog_items, proofs)

NOTIFICATIONS
    └── dealership_user_notification_prefs
    └── internal_user_notification_prefs
```

### Core Entities

#### Dealership Users

| Field | Type | Description |
|-------|------|-------------|
| id | serial | Internal PK |
| uuid | uuid | External identifier |
| dealership_id | int | FK to dealerships |
| name | text | Display name |
| email | citext | Unique email (case-insensitive) |
| avatar | text | Profile image URL |
| role | enum | viewer, submitter, approver, admin |
| is_active | bool | Soft delete flag |

**Roles:**
- `viewer`: Can see projects but not take action
- `submitter`: Can create projects, add inlays, chat
- `approver`: Can approve designs and place orders
- `admin`: Full access including user management

#### Internal Users

| Field | Type | Description |
|-------|------|-------------|
| id | serial | Internal PK |
| uuid | uuid | External identifier |
| name | text | Display name |
| email | citext | Unique email |
| avatar | text | Profile image URL |
| role | enum | designer, production, billing, admin |
| is_active | bool | Soft delete flag |

**Roles:**
- `designer`: Creates proofs, responds to design chats
- `production`: Manages kanban, creates blockers
- `billing`: Creates and manages invoices
- `admin`: Full access

#### Projects

| Field | Type | Description |
|-------|------|-------------|
| id | serial | Internal PK |
| uuid | uuid | External identifier |
| dealership_id | int | FK to dealerships |
| name | text | Project name |
| status | enum | Current lifecycle stage |
| ordered_at | timestamptz | When order was placed |
| ordered_by | int | FK to dealership_user who ordered |

**Status Flow:**
```
draft → designing → pending-approval → approved → ordered →
in-production → shipped → delivered → invoiced → completed

Any status → cancelled
```

#### Inlays

| Field | Type | Description |
|-------|------|-------------|
| id | serial | Internal PK |
| uuid | uuid | External identifier |
| project_id | int | FK to projects |
| name | text | Inlay name |
| type | enum | catalog or custom |
| preview_url | text | Current preview (auto-updated) |
| approved_proof_id | int | FK to approved proof |
| manufacturing_step | enum | Current kanban position |

**Manufacturing Steps:**
```
ordered → materials-prep → cutting → fire-polish → packaging → shipped → delivered
```

#### Inlay Proofs

| Field | Type | Description |
|-------|------|-------------|
| id | serial | Internal PK |
| uuid | uuid | External identifier |
| inlay_id | int | FK to inlays |
| version_number | int | Sequential version |
| preview_url | text | Proof image |
| design_asset_url | text | Production file (S3) |
| width | float | Final width |
| height | float | Final height |
| price_group_id | int | FK to price_groups |
| price_cents | int | Per-inlay price (future) |
| scale_factor | float | For graphical editor |
| color_overrides | jsonb | For graphical editor |
| status | enum | pending, approved, declined, superseded |
| approved_at | timestamptz | When approved |
| approved_by | int | FK to dealership_user |
| decline_reason | text | Feedback on decline |
| sent_in_chat_id | int | FK to chat message |

#### Inlay Chats

| Field | Type | Description |
|-------|------|-------------|
| id | serial | Internal PK |
| uuid | uuid | External identifier |
| inlay_id | int | FK to inlays |
| dealership_user_id | int | Sender (if dealership) |
| internal_user_id | int | Sender (if internal) |
| message_type | enum | text, image, proof_*, system |
| message | text | Message content |
| attachment_url | text | Image URL if applicable |

**Message Types:**
- `text`: Regular message
- `image`: Message with image attachment
- `proof_sent`: System message when proof shared
- `proof_approved`: System message on approval
- `proof_declined`: System message on decline
- `system`: Other automated messages

#### Inlay Milestones

| Field | Type | Description |
|-------|------|-------------|
| id | serial | Internal PK |
| uuid | uuid | External identifier |
| inlay_id | int | FK to inlays |
| step | enum | Manufacturing step |
| event_type | enum | entered, exited, reverted |
| performed_by | int | FK to internal_user |
| notes | text | Optional notes |
| event_time | timestamptz | When event occurred |

#### Inlay Blockers

| Field | Type | Description |
|-------|------|-------------|
| id | serial | Internal PK |
| uuid | uuid | External identifier |
| inlay_id | int | FK to inlays |
| blocker_type | enum | soft or hard |
| reason | text | Why blocked |
| step_blocked | text | Which step is blocked |
| created_by | int | FK to internal_user |
| resolved_at | timestamptz | When resolved |
| resolved_by | int | FK to internal_user |
| resolution_notes | text | How it was resolved |

**Blocker Types:**
- `soft`: Informational, doesn't prevent progress
- `hard`: Prevents moving to next step

#### Invoices

| Field | Type | Description |
|-------|------|-------------|
| id | serial | Internal PK |
| uuid | uuid | External identifier |
| project_id | int | FK to projects (1:1) |
| invoice_number | text | Human-readable number |
| subtotal_cents | int | Sum of line items |
| tax_cents | int | Tax amount |
| total_cents | int | Final total |
| status | enum | draft, sent, paid, void |
| sent_at | timestamptz | When emailed |
| paid_at | timestamptz | When payment received |

#### Order Snapshots

| Field | Type | Description |
|-------|------|-------------|
| id | serial | Internal PK |
| uuid | uuid | External identifier |
| project_id | int | FK to projects |
| inlay_id | int | FK to inlays (1:1) |
| proof_id | int | FK to approved proof |
| price_group_id | int | Locked price group |
| price_cents | int | Locked price |
| width | float | Locked dimensions |
| height | float | Locked dimensions |

---

## Feature Breakdown

### F1: Authentication & Authorization

#### F1.1 Dual User Authentication
- Separate token tables for dealership/internal users
- Middleware detects user type from token
- Context stores authenticated user with type

#### F1.2 Dealership User Permissions
- Role-based access control
- Centralized `<Can>` component for UI
- Permission checking utilities

#### F1.3 Internal User Permissions
- Role-based access control
- Separate permission set from dealership

### F2: Catalog Management

#### F2.1 Catalog CRUD (Internal)
- Create/edit catalog items
- Upload preview images
- Manage tags
- Set dimensions and pricing

#### F2.2 Catalog Browser (Dealership)
- Browse by category
- Filter by tags
- Search by code/name
- View item details

### F3: Project & Inlay Flow

#### F3.1 Project Management
- Create projects
- View project list with status
- Project detail view

#### F3.2 Inlay Management
- Add catalog inlays to project
- Add custom inlays with reference images
- Remove inlays from project
- View inlay details

#### F3.3 Order Placement
- Validate all inlays approved
- Create order snapshots
- Lock pricing
- Transition project status

### F4: Chat & Proofs

#### F4.1 Inlay Chat
- Real-time messaging per inlay
- Image attachments
- System messages

#### F4.2 Proof Workflow
- Create proofs (internal)
- Send proofs in chat
- Approve/decline proofs (dealership)
- Version history

#### F4.3 Project Chat
- Manufacturing-phase discussion
- Available after order placed

### F5: Manufacturing

#### F5.1 Kanban Board (Internal)
- View inlays by step
- Drag-and-drop movement
- Batch operations

#### F5.2 Milestone Tracking
- Record step transitions
- Support backward movement
- View history

#### F5.3 Blocker Management
- Create blockers
- Resolve blockers
- Notify dealership

### F6: Notifications

#### F6.1 Notification System
- Create notifications on events
- Store in database
- Send email

#### F6.2 In-App Notifications
- Notification list
- Unread count
- Mark as read
- Deep links

#### F6.3 Preferences
- Per-event-type settings
- Enable/disable email

### F7: Invoicing

#### F7.1 Invoice Management (Internal)
- Create from project
- Edit line items
- Send to dealership
- Mark as paid

#### F7.2 Invoice View (Dealership)
- View invoice details
- Download PDF

### F8: Dashboards

#### F8.1 Dealership Dashboard
- Projects needing action
- Recent activity
- Quick actions

#### F8.2 Internal Dashboard
- Kanban stats
- Pending items
- Alerts

---

## Implementation Phases

### Phase 1: Foundation (2 weeks)

**Goal:** Update data layer to match new schema

| Task | Est. | Dependencies |
|------|------|--------------|
| Run migrations, regenerate Jet | 2h | - |
| Update Go models for renamed tables | 4h | Migrations |
| Create new Go models | 16h | Migrations |
| Update TypeScript types | 8h | - |
| Update existing tests | 8h | Go models |
| Write tests for new models | 16h | Go models |

### Phase 2: Auth & Permissions (1 week)

**Goal:** Support dual user types with role-based permissions

| Task | Est. | Dependencies |
|------|------|--------------|
| Dual auth middleware | 8h | Phase 1 |
| Token system updates | 4h | Phase 1 |
| Permission utilities (Go) | 4h | Auth middleware |
| `<Can>` component | 4h | - |
| Permission hooks | 4h | - |
| Route protection | 4h | `<Can>` |
| Internal login flow | 8h | Auth middleware |

### Phase 3: Catalog System (1 week)

**Goal:** Complete catalog management and browsing

| Task | Est. | Dependencies |
|------|------|--------------|
| Catalog API endpoints | 8h | Phase 1 |
| Image upload to S3 | 4h | Phase 1 |
| Tag management API | 4h | Phase 1 |
| Internal catalog UI | 12h | Catalog API |
| Dealership catalog browser | 12h | Catalog API |

### Phase 4: Project & Inlay Flow (2 weeks)

**Goal:** Complete project creation and order placement

| Task | Est. | Dependencies |
|------|------|--------------|
| Update project API | 8h | Phase 1 |
| Update inlay API | 8h | Phase 1 |
| Order placement API | 12h | Phase 1 |
| Project creation UI | 16h | Project API |
| Inlay management UI | 12h | Inlay API |
| Order placement UI | 8h | Order API |
| Order snapshot creation | 8h | Order API |

### Phase 5: Chat & Proofs (1.5 weeks)

**Goal:** Design discussion and approval workflow

| Task | Est. | Dependencies |
|------|------|--------------|
| Update chat API | 8h | Phase 1 |
| Proof API | 12h | Phase 1 |
| Approve/decline API | 8h | Proof API |
| Chat UI refactor | 12h | Chat API |
| Proof display in chat | 8h | Proof API |
| Proof version history | 8h | Proof API |

### Phase 6: Manufacturing (1.5 weeks)

**Goal:** Kanban and blocker management

| Task | Est. | Dependencies |
|------|------|--------------|
| Kanban API | 8h | Phase 1 |
| Milestone API | 8h | Phase 1 |
| Blocker API | 8h | Phase 1 |
| Kanban board UI | 16h | Kanban API |
| Blocker management UI | 8h | Blocker API |
| Dealership progress view | 8h | Milestone API |

### Phase 7: Notifications (1 week)

**Goal:** Email notifications and in-app viewing

| Task | Est. | Dependencies |
|------|------|--------------|
| Notification service | 8h | Phase 1 |
| Email integration (SES) | 8h | Notification service |
| Event-to-notification mapping | 8h | All previous phases |
| Notification API | 4h | Notification service |
| Preferences API | 4h | Phase 1 |
| In-app notification UI | 12h | Notification API |

### Phase 8: Invoicing (1 week)

**Goal:** Invoice creation and management

| Task | Est. | Dependencies |
|------|------|--------------|
| Invoice API | 12h | Phase 1 |
| Invoice number generation | 2h | Invoice API |
| Internal invoice UI | 12h | Invoice API |
| Dealership invoice view | 8h | Invoice API |
| PDF generation | 8h | Invoice API |

### Phase 9: Dashboards (1 week)

**Goal:** Overview and quick actions

| Task | Est. | Dependencies |
|------|------|--------------|
| Dashboard queries | 8h | All previous |
| Dealership dashboard | 12h | Dashboard queries |
| Internal dashboard | 12h | Dashboard queries |
| Action item components | 8h | Dashboard queries |

---

## Task Dependencies

```
Phase 1 (Foundation)
    │
    ├──► Phase 2 (Auth)
    │        │
    │        └──► Phase 3 (Catalog)
    │                 │
    │                 └──► Phase 4 (Project/Inlay)
    │                          │
    │                          ├──► Phase 5 (Chat/Proofs)
    │                          │        │
    │                          │        └──► Phase 6 (Manufacturing)
    │                          │                 │
    │                          └─────────────────┴──► Phase 7 (Notifications)
    │                                                      │
    │                                                      └──► Phase 8 (Invoicing)
    │                                                               │
    └───────────────────────────────────────────────────────────────┴──► Phase 9 (Dashboards)
```

---

## MVP vs Post-MVP

### MVP Features

| Feature | Included | Notes |
|---------|----------|-------|
| Dual user authentication | ✅ | |
| Role-based permissions | ✅ | Preset roles only |
| Catalog management | ✅ | |
| Project creation | ✅ | |
| Inlay management | ✅ | Catalog + custom |
| Design chat | ✅ | Text + images |
| Proof workflow | ✅ | Version history |
| Order placement | ✅ | Price locking |
| Manufacturing kanban | ✅ | Fixed steps |
| Blockers | ✅ | Soft + hard |
| Email notifications | ✅ | |
| In-app notifications | ✅ | Polling |
| Invoicing | ✅ | Full payment only |
| Basic dashboards | ✅ | |

### Post-MVP Features

| Feature | Priority | Notes |
|---------|----------|-------|
| Graphical editor | High | Resize/recolor catalog items |
| SMS notifications | Medium | Twilio integration |
| Websocket notifications | Medium | Real-time updates |
| Shipping integration | Medium | UPS/FedEx webhooks |
| Granular permissions | Low | Per-action permissions |
| Partial payments | Low | Payment plans |
| Configurable kanban steps | Low | Dynamic workflow |
| Batch inlay operations | Low | Move multiple at once |
| Advanced reporting | Low | Revenue, turnaround time |
| Audit log viewer | Low | UI for existing audit data |

---

## TypeScript Types Summary

Types to create/update in `libs/data/src/`:

| File | Types | Status |
|------|-------|--------|
| `dealership-users.ts` | `DealershipUserRole`, `DealershipUser` | New |
| `internal-users.ts` | `InternalUserRole`, `InternalUser` | New |
| `price-groups.ts` | `PriceGroup` | New |
| `catalog-items.ts` | `CatalogItem`, `CatalogItemTag`, `CatalogItemImage` | Update |
| `projects.ts` | `ProjectStatus`, `Project` | Update |
| `inlays.ts` | `InlayType`, `ManufacturingStep`, `Inlay`, `InlayCatalogInfo`, `InlayCustomInfo` | Update |
| `inlay-chats.ts` | `ChatMessageType`, `InlayChat` | Update |
| `inlay-proofs.ts` | `ProofStatus`, `InlayProof` | Rewrite |
| `inlay-milestones.ts` | `MilestoneStep`, `MilestoneEventType`, `InlayMilestone` | Update |
| `inlay-blockers.ts` | `BlockerType`, `InlayBlocker` | New |
| `project-chats.ts` | `ProjectChat` | New |
| `order-snapshots.ts` | `OrderSnapshot` | New |
| `invoices.ts` | `InvoiceStatus`, `Invoice`, `InvoiceLineItem` | New |
| `notifications.ts` | `NotificationEventType`, `Notification`, `NotificationPreference` | New |

---

## Go Models Summary

Models to create/update in `libs/data/pkg/`:

| File | Model | Status |
|------|-------|--------|
| `dealership-users.go` | `DealershipUserModel` | Rename from users.go |
| `dealership-accounts.go` | `DealershipAccountModel` | Rename from accounts.go |
| `dealership-tokens.go` | `DealershipTokenModel` | Rename from tokens.go |
| `internal-users.go` | `InternalUserModel` | New |
| `internal-accounts.go` | `InternalAccountModel` | New |
| `internal-tokens.go` | `InternalTokenModel` | New |
| `price-groups.go` | `PriceGroupModel` | New |
| `catalog-items.go` | `CatalogItemModel` | New (was stub) |
| `projects.go` | `ProjectModel` | Update |
| `inlays.go` | `InlayModel` | Update |
| `inlay-chats.go` | `InlayChatModel` | Update |
| `inlay-proofs.go` | `InlayProofModel` | Rewrite |
| `inlay-milestones.go` | `InlayMilestoneModel` | New |
| `inlay-blockers.go` | `InlayBlockerModel` | New |
| `project-chats.go` | `ProjectChatModel` | New |
| `order-snapshots.go` | `OrderSnapshotModel` | New |
| `invoices.go` | `InvoiceModel` | New |
| `notifications.go` | `NotificationModel` | New |

---

## API Endpoints Summary

### Auth
- `GET /api/auth/google` - Google OAuth
- `GET /api/auth/google/callback` - Google callback
- `GET /api/auth/microsoft` - Microsoft OAuth
- `GET /api/auth/microsoft/callback` - Microsoft callback
- `POST /api/auth/magic-link` - Send magic link
- `GET /api/auth/magic-link/callback` - Magic link callback
- `POST /api/auth/token/access` - Refresh access token
- `GET /api/auth/logout` - Logout
- `GET /api/internal/auth/...` - Internal user auth routes

### Dealership Users
- `GET /api/dealership-user` - List users
- `GET /api/dealership-user/self` - Current user
- `GET /api/dealership-user/:uuid` - Get user
- `POST /api/dealership-user` - Create user
- `PATCH /api/dealership-user/:uuid` - Update user

### Internal Users
- `GET /api/internal-user` - List users
- `GET /api/internal-user/self` - Current user
- `GET /api/internal-user/:uuid` - Get user
- `POST /api/internal-user` - Create user
- `PATCH /api/internal-user/:uuid` - Update user

### Catalog
- `GET /api/catalog` - List items
- `GET /api/catalog/:uuid` - Get item
- `POST /api/catalog` - Create item
- `PATCH /api/catalog/:uuid` - Update item
- `POST /api/catalog/:uuid/images` - Upload image
- `DELETE /api/catalog/:uuid/images/:imageUuid` - Delete image
- `POST /api/catalog/:uuid/tags` - Add tag
- `DELETE /api/catalog/:uuid/tags/:tag` - Remove tag

### Projects
- `GET /api/project` - List projects
- `GET /api/project/:uuid` - Get project
- `POST /api/project` - Create project
- `PATCH /api/project/:uuid` - Update project
- `POST /api/project/:uuid/place-order` - Place order

### Inlays
- `GET /api/inlay` - List inlays
- `GET /api/inlay/:uuid` - Get inlay
- `POST /api/project/:uuid/inlays/catalog` - Add catalog inlay
- `POST /api/project/:uuid/inlays/custom` - Add custom inlay
- `DELETE /api/inlay/:uuid` - Remove inlay
- `GET /api/inlay/:uuid/proofs` - List proofs
- `GET /api/inlay/:uuid/milestones` - List milestones
- `GET /api/inlay/:uuid/blockers` - List blockers

### Proofs
- `POST /api/inlay/:uuid/proofs` - Create proof
- `GET /api/proof/:uuid` - Get proof
- `POST /api/proof/:uuid/approve` - Approve
- `POST /api/proof/:uuid/decline` - Decline

### Chats
- `GET /api/inlay/:uuid/chats` - Get inlay chat messages
- `POST /api/inlay/:uuid/chats` - Send message
- `GET /api/project/:uuid/chats` - Get project chat messages
- `POST /api/project/:uuid/chats` - Send message

### Manufacturing
- `GET /api/kanban` - Get kanban board
- `POST /api/inlay/:uuid/step` - Move to step
- `POST /api/inlay/:uuid/revert` - Revert step
- `POST /api/inlay/:uuid/blockers` - Create blocker
- `POST /api/blocker/:uuid/resolve` - Resolve blocker

### Invoices
- `GET /api/invoice` - List invoices
- `GET /api/invoice/:uuid` - Get invoice
- `POST /api/project/:uuid/invoice` - Create invoice
- `PATCH /api/invoice/:uuid` - Update invoice
- `POST /api/invoice/:uuid/line-items` - Add line item
- `PATCH /api/invoice/:uuid/line-items/:lineUuid` - Update line item
- `DELETE /api/invoice/:uuid/line-items/:lineUuid` - Remove line item
- `POST /api/invoice/:uuid/send` - Send invoice
- `POST /api/invoice/:uuid/mark-paid` - Mark paid

### Notifications
- `GET /api/notifications` - List notifications
- `GET /api/notifications/unread-count` - Get count
- `POST /api/notifications/:uuid/read` - Mark read
- `POST /api/notifications/read-all` - Mark all read
- `GET /api/notification-preferences` - Get preferences
- `PATCH /api/notification-preferences` - Update preferences
