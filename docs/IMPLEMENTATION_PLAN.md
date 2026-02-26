# GlassAct Studios - Implementation Plan

This document outlines the complete implementation plan for the GlassAct Studios ecommerce platform. It covers data models, feature breakdown, task dependencies, and development phases.

## Table of Contents

1. [Project Status](#project-status)
2. [Executive Summary](#executive-summary)
3. [Architecture Overview](#architecture-overview)
4. [Data Models](#data-models)
5. [Feature Breakdown](#feature-breakdown)
6. [Implementation Phases](#implementation-phases)
7. [Task Dependencies](#task-dependencies)
8. [MVP vs Post-MVP](#mvp-vs-post-mvp)

---

## Project Status

**Last Updated:** February 5, 2026

| Phase                             | Status      | Progress | Notes                                                           |
| --------------------------------- | ----------- | -------- | --------------------------------------------------------------- |
| **Phase 1: Foundation**           | ✅ COMPLETE | 100%     | Go models and TypeScript types complete.                        |
| **Phase 2: Auth & Permissions**   | ✅ COMPLETE | 100%     | Dual auth, unified OAuth, permissions, user management complete |
| **Phase 3: Catalog System**       | ✅ COMPLETE | 100%     | Admin CRUD, browsing, filtering, SVG upload complete            |
| **Phase 4: Project & Inlay Flow** | ⏳ Pending  | 0%       | Ready to start                           |
| **Phase 5: Chat & Proofs**        | ⏳ Pending  | 0%       | Ready to start                           |
| **Phase 6: Manufacturing**        | ⏳ Pending  | 0%       | Ready to start                           |
| **Phase 7: Notifications**        | ⏳ Pending  | 0%       | Ready to start                           |
| **Phase 8: Invoicing**            | ⏳ Pending  | 0%       | Ready to start                           |
| **Phase 9: Dashboards**           | ⏳ Pending  | 0%       | Ready to start                           |

---

## Executive Summary

GlassAct Studios is a B2B ecommerce platform for custom stained glass inlays in the memorial industry. The platform serves gravestone engravers ("dealerships") who order inlays on behalf of their customers.

### Key Differentiators

- **Conversational ordering**: Each inlay has its own design discussion thread
- **Non-linear manufacturing**: Inlays can move backward in the production process
- **Dual user systems**: Separate models for dealership users and internal staff
- **Proof-centric workflow**: Designs are versioned with explicit approval tracking

### Core Decisions

| Decision      | Choice                                  | Rationale                                     |
| ------------- | --------------------------------------- | --------------------------------------------- |
| User Model    | Separate tables for dealership/internal | Clean separation, different permission models |
| Permissions   | Role-based presets                      | Simpler MVP, centralized for future expansion |
| Manufacturing | Event-based milestones                  | Supports non-linear progression               |
| Proofs        | Versioned, in-chat                      | Natural conversation flow                     |
| Pricing       | Locked at order                         | Immutable snapshots for billing accuracy      |
| Notifications | Email only, polling                     | Simpler MVP, websockets later                 |

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

| Layer    | Technology                                               |
| -------- | -------------------------------------------------------- |
| Frontend | SolidJS, TanStack (Router, Query, Form, Table), Tailwind |
| Backend  | Go, standard library HTTP, Jet SQL builder               |
| Database | PostgreSQL with PostGIS                                  |
| Storage  | AWS S3 for images and design assets                      |
| Monorepo | pnpm workspaces                                          |

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

PRICE_GROUPS
    └── (referenced by catalog_items, proofs)

NOTIFICATIONS
    └── dealership_user_notification_prefs
    └── internal_user_notification_prefs
```

### Core Entities

#### Dealership Users

| Field         | Type   | Description                        |
| ------------- | ------ | ---------------------------------- |
| id            | serial | Internal PK                        |
| uuid          | uuid   | External identifier                |
| dealership_id | int    | FK to dealerships                  |
| name          | text   | Display name                       |
| email         | citext | Unique email (case-insensitive)    |
| avatar        | text   | Profile image URL                  |
| role          | enum   | viewer, submitter, approver, admin |
| is_active     | bool   | Soft delete flag                   |

**Roles:**

- `viewer`: Can see projects but not take action
- `submitter`: Can create projects, add inlays, chat
- `approver`: Can approve designs and place orders
- `admin`: Full access including user management

#### Internal Users

| Field     | Type   | Description                          |
| --------- | ------ | ------------------------------------ |
| id        | serial | Internal PK                          |
| uuid      | uuid   | External identifier                  |
| name      | text   | Display name                         |
| email     | citext | Unique email                         |
| avatar    | text   | Profile image URL                    |
| role      | enum   | designer, production, billing, admin |
| is_active | bool   | Soft delete flag                     |

**Roles:**

- `designer`: Creates proofs, responds to design chats
- `production`: Manages kanban, creates blockers
- `billing`: Creates and manages invoices
- `admin`: Full access

#### Projects

| Field         | Type        | Description                       |
| ------------- | ----------- | --------------------------------- |
| id            | serial      | Internal PK                       |
| uuid          | uuid        | External identifier               |
| dealership_id | int         | FK to dealerships                 |
| name          | text        | Project name                      |
| status        | enum        | Current lifecycle stage           |
| ordered_at    | timestamptz | When order was placed             |
| ordered_by    | int         | FK to dealership_user who ordered |

**Status Flow:**

```
draft → designing → pending-approval → approved → ordered →
in-production → shipped → delivered → invoiced → completed

Any status → cancelled
```

#### Inlays

| Field              | Type   | Description                    |
| ------------------ | ------ | ------------------------------ |
| id                 | serial | Internal PK                    |
| uuid               | uuid   | External identifier            |
| project_id         | int    | FK to projects                 |
| name               | text   | Inlay name                     |
| type               | enum   | catalog or custom              |
| preview_url        | text   | Current preview (auto-updated) |
| approved_proof_id  | int    | FK to approved proof           |
| manufacturing_step | enum   | Current kanban position        |

**Manufacturing Steps:**

```
ordered → materials-prep → cutting → fire-polish → packaging → shipped → delivered
```

#### Inlay Proofs

| Field            | Type        | Description                             |
| ---------------- | ----------- | --------------------------------------- |
| id               | serial      | Internal PK                             |
| uuid             | uuid        | External identifier                     |
| inlay_id         | int         | FK to inlays                            |
| version_number   | int         | Sequential version                      |
| preview_url      | text        | Proof image                             |
| design_asset_url | text        | Production file (S3)                    |
| width            | float       | Final width                             |
| height           | float       | Final height                            |
| price_group_id   | int         | FK to price_groups                      |
| price_cents      | int         | Per-inlay price (future)                |
| scale_factor     | float       | For graphical editor                    |
| color_overrides  | jsonb       | For graphical editor                    |
| status           | enum        | pending, approved, declined, superseded |
| approved_at      | timestamptz | When approved                           |
| approved_by      | int         | FK to dealership_user                   |
| decline_reason   | text        | Feedback on decline                     |
| sent_in_chat_id  | int         | FK to chat message                      |

#### Inlay Chats

| Field              | Type   | Description                    |
| ------------------ | ------ | ------------------------------ |
| id                 | serial | Internal PK                    |
| uuid               | uuid   | External identifier            |
| inlay_id           | int    | FK to inlays                   |
| dealership_user_id | int    | Sender (if dealership)         |
| internal_user_id   | int    | Sender (if internal)           |
| message_type       | enum   | text, image, proof\_\*, system |
| message            | text   | Message content                |
| attachment_url     | text   | Image URL if applicable        |

**Message Types:**

- `text`: Regular message
- `image`: Message with image attachment
- `proof_sent`: System message when proof shared
- `proof_approved`: System message on approval
- `proof_declined`: System message on decline
- `system`: Other automated messages

#### Inlay Milestones

| Field        | Type        | Description               |
| ------------ | ----------- | ------------------------- |
| id           | serial      | Internal PK               |
| uuid         | uuid        | External identifier       |
| inlay_id     | int         | FK to inlays              |
| step         | enum        | Manufacturing step        |
| event_type   | enum        | entered, exited, reverted |
| performed_by | int         | FK to internal_user       |
| notes        | text        | Optional notes            |
| event_time   | timestamptz | When event occurred       |

#### Inlay Blockers

| Field            | Type        | Description           |
| ---------------- | ----------- | --------------------- |
| id               | serial      | Internal PK           |
| uuid             | uuid        | External identifier   |
| inlay_id         | int         | FK to inlays          |
| blocker_type     | enum        | soft or hard          |
| reason           | text        | Why blocked           |
| step_blocked     | text        | Which step is blocked |
| created_by       | int         | FK to internal_user   |
| resolved_at      | timestamptz | When resolved         |
| resolved_by      | int         | FK to internal_user   |
| resolution_notes | text        | How it was resolved   |

**Blocker Types:**

- `soft`: Informational, doesn't prevent progress
- `hard`: Prevents moving to next step

#### Invoices

| Field          | Type        | Description             |
| -------------- | ----------- | ----------------------- |
| id             | serial      | Internal PK             |
| uuid           | uuid        | External identifier     |
| project_id     | int         | FK to projects (1:1)    |
| invoice_number | text        | Human-readable number   |
| subtotal_cents | int         | Sum of line items       |
| tax_cents      | int         | Tax amount              |
| total_cents    | int         | Final total             |
| status         | enum        | draft, sent, paid, void |
| sent_at        | timestamptz | When emailed            |
| paid_at        | timestamptz | When payment received   |

#### Order Snapshots

| Field          | Type   | Description          |
| -------------- | ------ | -------------------- |
| id             | serial | Internal PK          |
| uuid           | uuid   | External identifier  |
| project_id     | int    | FK to projects       |
| inlay_id       | int    | FK to inlays (1:1)   |
| proof_id       | int    | FK to approved proof |
| price_group_id | int    | Locked price group   |
| price_cents    | int    | Locked price         |
| width          | float  | Locked dimensions    |
| height         | float  | Locked dimensions    |

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

### Phase 1: Foundation

**Goal:** Update data layer to match new schema

**Status:** ✅ **COMPLETE** (Feb 4, 2026)

#### Completed Tasks

| Task                                | Status                                   |
| ----------------------------------- | ---------------------------------------- |
| Run migrations, regenerate Jet      | ✅ Complete                              |
| Update Go models for renamed tables | ✅ Complete (dealership*\*, internal*\*) |
| Create new Go models                | ✅ Complete (18 total models)            |
| Update TypeScript types             | ✅ Complete                              |
| Update existing tests               | ✅ Complete                              |
| Write tests for new models          | ✅ Complete                              |

#### Phase 1 Deliverables

**18 Go Models Created/Updated:**

**Stage 1: User & Auth (6 models)**

- `dealership_users.go` - DealershipUser, DealershipUserModel, DealershipUserRole (viewer, submitter, approver, admin)
- `dealership_accounts.go` - DealershipAccount, DealershipAccountModel
- `dealership_tokens.go` - DealershipToken, DealershipTokenModel
- `internal_users.go` - InternalUser, InternalUserModel, InternalUserRole (designer, production, billing, admin)
- `internal_accounts.go` - InternalAccount, InternalAccountModel
- `internal_tokens.go` - InternalToken, InternalTokenModel

**Stage 2: Foundational Business (2 models)**

- `price_groups.go` - PriceGroup, PriceGroupModel
- `catalog_items.go` - CatalogItem, CatalogItemTag, CatalogItemModel (with tag management)

**Stage 3: Core Business (4 models)**

- `projects.go` - Project, ProjectStatus, ProjectModel (with new status flow)
- `invoices.go` - Invoice, InvoiceLineItem, InvoiceStatus, InvoiceModel
- `order_snapshots.go` - OrderSnapshot, OrderSnapshotModel
- `inlays.go` - Inlay, InlayCatalogInfo, InlayCustomInfo, InlayType, InlayModel

**Stage 4: Discussion & Tracking (6 models)**

- `inlay_chats.go` - InlayChat, ChatMessageType, InlayChatModel
- `project_chats.go` - ProjectChat, ProjectChatModel
- `inlay_proofs.go` - InlayProof, ProofStatus, InlayProofModel (with JSON color overrides)
- `inlay_milestones.go` - InlayMilestone, ManufacturingStep, MilestoneEventType, InlayMilestoneModel
- `inlay_blockers.go` - InlayBlocker, BlockerType, InlayBlockerModel
- `notifications.go` - Notification, NotificationEventType, NotificationModel (with unread tracking)

**Implementation Details:**

- All models follow Jet ORM integration pattern
- Standard FromGen/ToGen conversion functions
- CRUD operations with 3-second context timeouts
- Pointer handling for nullable fields
- UUID support throughout
- Specialized query methods (GetByTag, GetUnread, GetApproved, etc.)
- JSON marshaling for complex types
- Underscore naming convention for all files
- Removed old files: accounts.go, tokens.go, users.go, inlay-chats.go (hyphenated versions)

**models.go Updated:**

- Registered all 18 models in Models struct
- Updated NewModels() factory function
- Alphabetically organized for maintainability

### Phase 2: Auth & Permissions

**Status:** ✅ **COMPLETE** (Feb 5, 2026)

**Goal:** Support dual user types with role-based permissions

#### Completed Features

**Backend Auth System:**
- ✅ `AuthUser` interface implemented on both `DealershipUser` and `InternalUser`
- ✅ Unified middleware using `GetAuthUserForToken()` for both user types
- ✅ Generic `Can(action string)` permission system with 13 permission actions
- ✅ Context helpers: `ContextSetAuthUser()`, `ContextGetDealershipUser()`, `ContextGetInternalUser()`
- ✅ Permission utilities: `RequirePermission()` and `RequireRole()` middleware

**Invite-Only Authentication:**
- ✅ Unified OAuth callback checks both `dealership_users` and `internal_users` tables
- ✅ Magic link flow supports both user types
- ✅ Returns 401 if user not pre-registered in either table
- ✅ Token refresh uses unified lookup with scope mapping

**User Management APIs:**
- ✅ POST/PATCH/DELETE `/api/dealership-user` (admin only)
- ✅ POST/PATCH/DELETE `/api/internal-user` (admin only)
- ✅ Dealership admins can only manage users in their dealership
- ✅ Internal admins can manage all internal users

**Frontend Auth System:**
- ✅ Auth union type: `type User = DealershipUser | InternalUser`
- ✅ Type guards: `isDealershipUser()`, `isInternalUser()`
- ✅ Auth context with: `isDealership()`, `isInternal()`, `can(action)` helpers
- ✅ `<Can>` component for permission-based conditional rendering
- ✅ Permission constants export for consistency

**OAuth Integration:**
- ✅ Same OAuth flows (Google, Microsoft) for both user types
- ✅ OAuth callback queries dealership first, then internal users
- ✅ Automatic account linking for existing emails
- ✅ `is_active` field enforcement for all auth flows

#### Permission Actions

**Dealership:**
- `create_project` - submitter, approver, admin
- `approve_proof` - approver, admin
- `place_order` - approver, admin
- `pay_invoice` - admin only
- `manage_dealership_users` - admin only
- `view_projects` - all roles
- `view_invoices` - all roles

**Internal:**
- `create_proof` - designer, admin
- `manage_kanban` - production, admin
- `create_blocker` - production, admin
- `create_invoice` - billing, admin
- `manage_internal_users` - admin only
- `view_all` - admin only

#### Files Created/Modified

**Backend:**
- `libs/data/pkg/auth.go` - AuthUser interface, scope constants
- `libs/data/pkg/permissions.go` - Permission action constants
- `libs/data/pkg/token_lookup.go` - Unified token lookup with scope mapping
- `libs/data/pkg/dealership_users.go` - Added AuthUser implementation, Can() method
- `libs/data/pkg/internal_users.go` - Added AuthUser implementation, Can() method
- `apps/api/app/context.go` - Updated to use AuthUser interface
- `apps/api/app/permissions.go` - RequirePermission and RequireRole middleware
- `apps/api/app/errors.go` - Added Forbidden error type
- `apps/api/modules/auth/authHandlers.go` - Updated all handlers for dual auth
- `apps/api/modules/auth/authServices.go` - Unified login, user lookup for both types
- `apps/api/modules/user/userHandlers.go` - Added CRUD endpoints for user management
- `apps/api/modules/modules.go` - Updated routes for new endpoints

**Frontend:**
- `libs/data/src/auth.ts` - Union type, type guards, permission constants
- `libs/data/src/index.ts` - Export auth module
- `apps/webapp/src/providers/user.tsx` - Updated with permission checking
- `apps/webapp/src/components/Can.tsx` - Permission-based rendering component

**Documentation:**
- `.cursor/rules/backend.md` - Added auth usage examples

### Phase 3: Catalog System

**Status:** ✅ **COMPLETE** (Feb 5, 2026)

**Goal:** Complete catalog management and browsing

#### Completed Features

**Backend API Endpoints:**
- ✅ `GET /api/catalog` - List items with pagination, filtering (name, code, category, active status)
- ✅ `POST /api/catalog` - Create new catalog item
- ✅ `GET /api/catalog/:uuid` - Get single item details
- ✅ `PATCH /api/catalog/:uuid` - Update catalog item (partial updates supported)
- ✅ `DELETE /api/catalog/:uuid` - Soft delete catalog item
- ✅ `POST /api/catalog/:uuid/tags` - Add tag to item
- ✅ `DELETE /api/catalog/:uuid/tags/:tag` - Remove tag from item
- ✅ `GET /api/catalog/browse` - Public catalog browsing with multi-criteria filtering
- ✅ `GET /api/catalog/categories` - Get distinct categories from active items
- ✅ `GET /api/catalog/tags` - Get all available tags

**Admin-Side Features:**

**Admin Catalog List Page** (`admin.catalog.tsx`)
- ✅ Full-featured data table with TanStack Solid Table
- ✅ Sortable and filterable columns
- ✅ Search by code and name
- ✅ Category display with dimensions
- ✅ Active/Inactive status badges
- ✅ Pagination (50 items per page, max 100)
- ✅ Toggle to include/exclude inactive items
- ✅ Edit and delete actions per item

**Admin Create Catalog Item Page** (`admin.catalog_.create.tsx`)
- ✅ Dedicated creation route
- ✅ Form validation and error handling
- ✅ Success redirection
- ✅ Tag creation on item creation
- ✅ Back navigation button

**Admin Edit Catalog Item Page** (`admin.catalog_.$uuid.tsx`)
- ✅ Dynamic routing by UUID
- ✅ Load full item details
- ✅ Update item properties
- ✅ Add and remove tags
- ✅ Delete entire item with confirmation
- ✅ Tag synchronization and validation
- ✅ Error state handling

**Catalog Form Component** (`/apps/webapp/src/components/admin/catalog-form.tsx`)
- ✅ Basic fields: catalog code, name, description, category, active status
- ✅ Dimensions: default and minimum width/height with validation
- ✅ Pricing: price group selection via combobox
- ✅ SVG asset upload (file upload integration)
- ✅ Tag management: input with autocomplete, suggestions from existing tags
- ✅ Comprehensive Zod schema validation
- ✅ Dynamic price group loading
- ✅ Duplicate tag prevention

**Customer-Facing Features:**

**Public Catalog Browse Page** (`/apps/webapp/src/routes/_app/catalog.index.tsx`)
- ✅ Two-panel layout: filter sidebar + responsive grid
- ✅ Responsive design (1 col mobile, 2 col tablet, 3 col desktop)
- ✅ Infinite scroll with "Load More" button
- ✅ Item cards showing: SVG preview, code, name, category
- ✅ Loading skeletons and empty states

**Filter Sidebar Component** (`/apps/webapp/src/components/catalog/filter-sidebar.tsx`)
- ✅ Search input (searches name and code)
- ✅ Category dropdown (populated from API)
- ✅ Tag multi-select with autocomplete
- ✅ Clear all filters button
- ✅ Dynamic filter suggestions

**Catalog Grid Component** (`/apps/webapp/src/components/catalog/catalog-grid.tsx`)
- ✅ Responsive grid layout with proper spacing
- ✅ Item cards with SVG preview images
- ✅ Catalog code badge (monospace)
- ✅ "View Details" button per item
- ✅ Loading and empty states

**Item Detail Modal** (`/apps/webapp/src/components/catalog/item-detail-modal.tsx`)
- ✅ Full-screen modal dialog
- ✅ Large SVG preview
- ✅ Complete item information display
- ✅ Tags display
- ✅ Scrollable for small screens

**Data Layer & Models:**
- ✅ `CatalogItem` and `CatalogItemTag` Go models with full CRUD
- ✅ Comprehensive query methods: GetByCode, GetByTag, GetActive, etc.
- ✅ Tag management in data layer
- ✅ Version-based optimistic locking
- ✅ TypeScript types aligned with Go models

**Integration Points:**
- ✅ Price group integration (select during item creation/editing)
- ✅ SVG file upload via `/api/upload` endpoint
- ✅ File storage with "catalog-items" path
- ✅ Soft delete implementation (is_active flag)
- ✅ Permission-based access (admin for write, authenticated for read)

#### Permission Controls

- ✅ Admin-only CRUD endpoints (`RequireRole("admin")` middleware)
- ✅ Public browse endpoints (requires authentication)
- ✅ Category and tag lists (requires authentication)

### Phase 4: Project & Inlay Flow

**Goal:** Complete project creation and order placement

| Task                    | Dependencies |
| ----------------------- | ------------ |
| Update project API      | Phase 1      |
| Update inlay API        | Phase 1      |
| Order placement API     | Phase 1      |
| Project creation UI     | Project API  |
| Inlay management UI     | Inlay API    |
| Order placement UI      | Order API    |
| Order snapshot creation | Order API    |

### Phase 5: Chat & Proofs

**Goal:** Design discussion and approval workflow

| Task                  | Dependencies |
| --------------------- | ------------ |
| Update chat API       | Phase 1      |
| Proof API             | Phase 1      |
| Approve/decline API   | Proof API    |
| Chat UI refactor      | Chat API     |
| Proof display in chat | Proof API    |
| Proof version history | Proof API    |

### Phase 6: Manufacturing

**Goal:** Kanban and blocker management

| Task                     | Dependencies  |
| ------------------------ | ------------- |
| Kanban API               | Phase 1       |
| Milestone API            | Phase 1       |
| Blocker API              | Phase 1       |
| Kanban board UI          | Kanban API    |
| Blocker management UI    | Blocker API   |
| Dealership progress view | Milestone API |

### Phase 7: Notifications

**Goal:** Email notifications and in-app viewing

| Task                          | Dependencies         |
| ----------------------------- | -------------------- |
| Notification service          | Phase 1              |
| Email integration (SES)       | Notification service |
| Event-to-notification mapping | All previous phases  |
| Notification API              | Notification service |
| Preferences API               | Phase 1              |
| In-app notification UI        | Notification API     |

### Phase 8: Invoicing

**Goal:** Invoice creation and management

| Task                      | Dependencies |
| ------------------------- | ------------ |
| Invoice API               | Phase 1      |
| Invoice number generation | Invoice API  |
| Internal invoice UI       | Invoice API  |
| Dealership invoice view   | Invoice API  |
| PDF generation            | Invoice API  |

### Phase 9: Dashboards

**Goal:** Overview and quick actions

| Task                   | Dependencies      |
| ---------------------- | ----------------- |
| Dashboard queries      | All previous      |
| Dealership dashboard   | Dashboard queries |
| Internal dashboard     | Dashboard queries |
| Action item components | Dashboard queries |

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

| Feature                  | Included | Notes             |
| ------------------------ | -------- | ----------------- |
| Dual user authentication | ✅       |                   |
| Role-based permissions   | ✅       | Preset roles only |
| Catalog management       | ✅       |                   |
| Project creation         | ✅       |                   |
| Inlay management         | ✅       | Catalog + custom  |
| Design chat              | ✅       | Text + images     |
| Proof workflow           | ✅       | Version history   |
| Order placement          | ✅       | Price locking     |
| Manufacturing kanban     | ✅       | Fixed steps       |
| Blockers                 | ✅       | Soft + hard       |
| Email notifications      | ✅       |                   |
| In-app notifications     | ✅       | Polling           |
| Invoicing                | ✅       | Full payment only |
| Basic dashboards         | ✅       |                   |

### Post-MVP Features

| Feature                   | Priority | Notes                        |
| ------------------------- | -------- | ---------------------------- |
| Graphical editor          | High     | Resize/recolor catalog items |
| SMS notifications         | Medium   | Twilio integration           |
| Websocket notifications   | Medium   | Real-time updates            |
| Shipping integration      | Medium   | UPS/FedEx webhooks           |
| Granular permissions      | Low      | Per-action permissions       |
| Partial payments          | Low      | Payment plans                |
| Configurable kanban steps | Low      | Dynamic workflow             |
| Batch inlay operations    | Low      | Move multiple at once        |
| Advanced reporting        | Low      | Revenue, turnaround time     |
| Audit log viewer          | Low      | UI for existing audit data   |

---

## TypeScript Types Summary

Types to create/update in `libs/data/src/`:

| File                  | Types                                                                            | Status  |
| --------------------- | -------------------------------------------------------------------------------- | ------- |
| `dealership-users.ts` | `DealershipUserRole`, `DealershipUser`                                           | New     |
| `internal-users.ts`   | `InternalUserRole`, `InternalUser`                                               | New     |
| `price-groups.ts`     | `PriceGroup`                                                                     | New     |
| `catalog-items.ts`    | `CatalogItem`, `CatalogItemTag`                                                  | Update  |
| `projects.ts`         | `ProjectStatus`, `Project`                                                       | Update  |
| `inlays.ts`           | `InlayType`, `ManufacturingStep`, `Inlay`, `InlayCatalogInfo`, `InlayCustomInfo` | Update  |
| `inlay-chats.ts`      | `ChatMessageType`, `InlayChat`                                                   | Update  |
| `inlay-proofs.ts`     | `ProofStatus`, `InlayProof`                                                      | Rewrite |
| `inlay-milestones.ts` | `MilestoneStep`, `MilestoneEventType`, `InlayMilestone`                          | Update  |
| `inlay-blockers.ts`   | `BlockerType`, `InlayBlocker`                                                    | New     |
| `project-chats.ts`    | `ProjectChat`                                                                    | New     |
| `order-snapshots.ts`  | `OrderSnapshot`                                                                  | New     |
| `invoices.ts`         | `InvoiceStatus`, `Invoice`, `InvoiceLineItem`                                    | New     |
| `notifications.ts`    | `NotificationEventType`, `Notification`, `NotificationPreference`                | New     |

---

## Go Models Summary

Models created/updated in `libs/data/pkg/` (Phase 1 Complete):

| File                     | Model                    | Status                                       | Completed   |
| ------------------------ | ------------------------ | -------------------------------------------- | ----------- |
| `dealership_users.go`    | `DealershipUserModel`    | ✅ Complete (renamed from users.go)          | Feb 4, 2026 |
| `dealership_accounts.go` | `DealershipAccountModel` | ✅ Complete (renamed from accounts.go)       | Feb 4, 2026 |
| `dealership_tokens.go`   | `DealershipTokenModel`   | ✅ Complete (renamed from tokens.go)         | Feb 4, 2026 |
| `internal_users.go`      | `InternalUserModel`      | ✅ Complete (new)                            | Feb 4, 2026 |
| `internal_accounts.go`   | `InternalAccountModel`   | ✅ Complete (new)                            | Feb 4, 2026 |
| `internal_tokens.go`     | `InternalTokenModel`     | ✅ Complete (new)                            | Feb 4, 2026 |
| `price_groups.go`        | `PriceGroupModel`        | ✅ Complete (new)                            | Feb 4, 2026 |
| `catalog_items.go`       | `CatalogItemModel`       | ✅ Complete (new, with tag management)       | Feb 4, 2026 |
| `projects.go`            | `ProjectModel`           | ✅ Complete (rewritten with new status flow) | Feb 4, 2026 |
| `inlays.go`              | `InlayModel`             | ✅ Complete (updated)                        | Feb 4, 2026 |
| `inlay_chats.go`         | `InlayChatModel`         | ✅ Complete (updated with message types)     | Feb 4, 2026 |
| `inlay_proofs.go`        | `InlayProofModel`        | ✅ Complete (rewritten with JSON overrides)  | Feb 4, 2026 |
| `inlay_milestones.go`    | `InlayMilestoneModel`    | ✅ Complete (new)                            | Feb 4, 2026 |
| `inlay_blockers.go`      | `InlayBlockerModel`      | ✅ Complete (new)                            | Feb 4, 2026 |
| `project_chats.go`       | `ProjectChatModel`       | ✅ Complete (new)                            | Feb 4, 2026 |
| `order_snapshots.go`     | `OrderSnapshotModel`     | ✅ Complete (new)                            | Feb 4, 2026 |
| `invoices.go`            | `InvoiceModel`           | ✅ Complete (new, with line items)           | Feb 4, 2026 |
| `notifications.go`       | `NotificationModel`      | ✅ Complete (new, with unread tracking)      | Feb 4, 2026 |

**Note:** All files use underscore naming convention (Go convention). Old hyphenated files have been removed.

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
