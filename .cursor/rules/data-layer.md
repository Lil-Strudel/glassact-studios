# GlassAct Studios - Data Layer Rules

These rules apply to the shared data layer in `libs/data/`.

## Overview

The `libs/data/` package serves two purposes:

1. **TypeScript types** (`src/`) - Shared types for frontend and type-safe API contracts
2. **Go data layer** (`pkg/`) - Database models and Jet-based queries

## TypeScript Types

### Location

All types live in `libs/data/src/`, one file per entity:

```
libs/data/src/
├── index.ts              # Re-exports everything
├── helpers.ts            # Type utilities
├── dealership-users.ts
├── internal-users.ts
├── projects.ts
├── inlays.ts
├── inlay-proofs.ts
└── ...
```

### The StandardTable Pattern

Most entities use `StandardTable<T>` which adds common fields:

```typescript
// helpers.ts
interface Metadata {
  created_at: string;
  updated_at: string;
  version: number;
}

interface DoubleId {
  id: number;
  uuid: string;
}

export type hasMetadata<T> = Tagged<T, "Metadata", Metadata>;
export type hasDoubleId<T> = Tagged<T, "DoubleId", DoubleId>;
export type StandardTable<T> = hasMetadata<hasDoubleId<T>>;
```

**Usage:**

```typescript
// projects.ts
import { StandardTable } from "./helpers";

export type ProjectStatus =
  | "draft"
  | "designing"
  | "pending-approval"
  | "approved"
  | "ordered"
  | "in-production"
  | "shipped"
  | "delivered"
  | "invoiced"
  | "completed"
  | "cancelled";

export type Project = StandardTable<{
  name: string;
  status: ProjectStatus;
  dealership_id: number;
  ordered_at: string | null;
  ordered_by: number | null;
}>;
```

### Type Helpers for API Operations

```typescript
// GET<T> - Full entity as returned from API
// Includes id, uuid, created_at, updated_at, version
type ProjectResponse = GET<Project>;
// Result: {
//   id: number;
//   uuid: string;
//   name: string;
//   status: ProjectStatus;
//   dealership_id: number;
//   ordered_at: string | null;
//   ordered_by: number | null;
//   created_at: string;
//   updated_at: string;
//   version: number;
// }

// POST<T> - Create request body
// Excludes id, uuid, created_at, updated_at, version
type CreateProjectRequest = POST<Project>;
// Result: {
//   name: string;
//   status: ProjectStatus;
//   dealership_id: number;
//   ordered_at: string | null;
//   ordered_by: number | null;
// }

// PATCH<T> - Update request body
// All fields optional except id or uuid
type UpdateProjectRequest = PATCH<Project>;
// Result: {
//   id?: number;
//   uuid?: string;
//   name?: string;
//   status?: ProjectStatus;
//   ...
// }
```

### Enum Pattern

Define string literal unions with a constant array for runtime validation:

```typescript
export type ProjectStatus =
  | "draft"
  | "designing"
  | "pending-approval"
  | "approved"
  | "ordered"
  | "in-production"
  | "shipped"
  | "delivered"
  | "invoiced"
  | "completed"
  | "cancelled";

export const PROJECT_STATUSES: ProjectStatus[] = [
  "draft",
  "designing",
  "pending-approval",
  "approved",
  "ordered",
  "in-production",
  "shipped",
  "delivered",
  "invoiced",
  "completed",
  "cancelled",
];
```

### Nested Types

For entities with sub-types (like Inlay with CatalogInfo or CustomInfo):

```typescript
export type InlayCatalogInfo = StandardTable<{
  inlay_id: number;
  catalog_item_id: number;
  customization_notes: string | null;
}>;

export type InlayCustomInfo = StandardTable<{
  inlay_id: number;
  description: string;
  requested_width: number | null;
  requested_height: number | null;
}>;

export type InlayType = "catalog" | "custom";

export type ManufacturingStep =
  | "ordered"
  | "materials-prep"
  | "cutting"
  | "fire-polish"
  | "packaging"
  | "shipped"
  | "delivered";

export type Inlay = StandardTable<{
  project_id: number;
  name: string;
  type: InlayType;
  preview_url: string;
  approved_proof_id: number | null;
  manufacturing_step: ManufacturingStep | null;
}>;

// For API responses that include nested data
export type InlayWithInfo = GET<Inlay> & {
  catalog_info?: GET<InlayCatalogInfo>;
  custom_info?: GET<InlayCustomInfo>;
};
```

### Nullable Fields

Use `| null` for nullable database columns:

```typescript
export type InlayProof = StandardTable<{
  inlay_id: number;
  version_number: number;
  preview_url: string;
  design_asset_url: string | null; // Optional S3 URL
  price_group_id: number | null; // Set when proof is sent
  price_cents: number | null; // Future per-inlay pricing
  approved_at: string | null; // Set when approved
  approved_by: number | null; // FK to dealership_user
  decline_reason: string | null; // Set when declined
}>;
```

### Index File

Export everything from `index.ts`:

```typescript
// index.ts
export * from "./helpers";
export * from "./dealership-users";
export * from "./internal-users";
export * from "./projects";
export * from "./inlays";
export * from "./inlay-proofs";
// ... etc
```

## Go Data Layer

### Location

```
libs/data/pkg/
├── gen/                  # Jet-generated code (don't edit)
│   └── glassact/
│       └── public/
│           ├── model/    # Generated structs
│           └── table/    # Generated table references
├── helpers.go            # Utilities
├── models.go             # Models registry
├── pool.go               # Connection pool
├── dealership-users.go
├── internal-users.go
├── projects.go
└── ...
```

### StandardTable Struct

```go
// helpers.go
type StandardTable struct {
    ID        int       `json:"id"`
    UUID      string    `json:"uuid"`
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
    Version   int       `json:"version"`
}
```

### Model Pattern

Each entity follows this pattern:

```go
// projects.go
package data

import (
    "context"
    "database/sql"
    "errors"
    "time"

    "github.com/Lil-Strudel/glassact-studios/libs/data/pkg/gen/glassact/public/model"
    "github.com/Lil-Strudel/glassact-studios/libs/data/pkg/gen/glassact/public/table"
    "github.com/go-jet/jet/v2/postgres"
    "github.com/go-jet/jet/v2/qrm"
    "github.com/google/uuid"
    "github.com/jackc/pgx/v5/pgxpool"
)

// 1. Status/enum type
type ProjectStatus string

type projectStatuses struct {
    Draft           ProjectStatus
    Designing       ProjectStatus
    PendingApproval ProjectStatus
    Approved        ProjectStatus
    Ordered         ProjectStatus
    InProduction    ProjectStatus
    Shipped         ProjectStatus
    Delivered       ProjectStatus
    Invoiced        ProjectStatus
    Completed       ProjectStatus
    Cancelled       ProjectStatus
}

var ProjectStatuses = projectStatuses{
    Draft:           ProjectStatus("draft"),
    Designing:       ProjectStatus("designing"),
    PendingApproval: ProjectStatus("pending-approval"),
    Approved:        ProjectStatus("approved"),
    Ordered:         ProjectStatus("ordered"),
    InProduction:    ProjectStatus("in-production"),
    Shipped:         ProjectStatus("shipped"),
    Delivered:       ProjectStatus("delivered"),
    Invoiced:        ProjectStatus("invoiced"),
    Completed:       ProjectStatus("completed"),
    Cancelled:       ProjectStatus("cancelled"),
}

// 2. API struct (what handlers use)
type Project struct {
    StandardTable
    Name         string        `json:"name"`
    Status       ProjectStatus `json:"status"`
    DealershipID int           `json:"dealership_id"`
    OrderedAt    *time.Time    `json:"ordered_at"`
    OrderedBy    *int          `json:"ordered_by"`
}

// 3. Model struct
type ProjectModel struct {
    DB   *pgxpool.Pool
    STDB *sql.DB
}

// 4. Conversion from Jet model
func projectFromGen(gen model.Projects) *Project {
    project := &Project{
        StandardTable: StandardTable{
            ID:        int(gen.ID),
            UUID:      gen.UUID.String(),
            CreatedAt: gen.CreatedAt,
            UpdatedAt: gen.UpdatedAt,
            Version:   int(gen.Version),
        },
        Name:         gen.Name,
        Status:       ProjectStatus(gen.Status),
        DealershipID: int(gen.DealershipID),
    }

    if gen.OrderedAt != nil {
        project.OrderedAt = gen.OrderedAt
    }
    if gen.OrderedBy != nil {
        orderedBy := int(*gen.OrderedBy)
        project.OrderedBy = &orderedBy
    }

    return project
}

// 5. Conversion to Jet model
func projectToGen(p *Project) (*model.Projects, error) {
    var projectUUID uuid.UUID
    var err error

    if p.UUID != "" {
        projectUUID, err = uuid.Parse(p.UUID)
        if err != nil {
            return nil, err
        }
    }

    gen := &model.Projects{
        ID:           int32(p.ID),
        UUID:         projectUUID,
        Name:         p.Name,
        Status:       string(p.Status),
        DealershipID: int32(p.DealershipID),
        CreatedAt:    p.CreatedAt,
        UpdatedAt:    p.UpdatedAt,
        Version:      int32(p.Version),
    }

    if p.OrderedAt != nil {
        gen.OrderedAt = p.OrderedAt
    }
    if p.OrderedBy != nil {
        orderedBy := int32(*p.OrderedBy)
        gen.OrderedBy = &orderedBy
    }

    return gen, nil
}

// 6. CRUD methods
func (m ProjectModel) Insert(project *Project) error {
    gen, err := projectToGen(project)
    if err != nil {
        return err
    }

    query := table.Projects.INSERT(
        table.Projects.Name,
        table.Projects.Status,
        table.Projects.DealershipID,
    ).MODEL(gen).RETURNING(
        table.Projects.ID,
        table.Projects.UUID,
        table.Projects.CreatedAt,
        table.Projects.UpdatedAt,
        table.Projects.Version,
    )

    ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
    defer cancel()

    var dest model.Projects
    err = query.QueryContext(ctx, m.STDB, &dest)
    if err != nil {
        return err
    }

    project.ID = int(dest.ID)
    project.UUID = dest.UUID.String()
    project.CreatedAt = dest.CreatedAt
    project.UpdatedAt = dest.UpdatedAt
    project.Version = int(dest.Version)

    return nil
}

// Transaction variant
func (m ProjectModel) TxInsert(tx *sql.Tx, project *Project) error {
    // Same as Insert but uses tx instead of m.STDB
    gen, err := projectToGen(project)
    if err != nil {
        return err
    }

    query := table.Projects.INSERT(
        table.Projects.Name,
        table.Projects.Status,
        table.Projects.DealershipID,
    ).MODEL(gen).RETURNING(
        table.Projects.ID,
        table.Projects.UUID,
        table.Projects.CreatedAt,
        table.Projects.UpdatedAt,
        table.Projects.Version,
    )

    ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
    defer cancel()

    var dest model.Projects
    err = query.QueryContext(ctx, tx, &dest) // Use tx here
    if err != nil {
        return err
    }

    project.ID = int(dest.ID)
    project.UUID = dest.UUID.String()
    project.CreatedAt = dest.CreatedAt
    project.UpdatedAt = dest.UpdatedAt
    project.Version = int(dest.Version)

    return nil
}

func (m ProjectModel) GetByID(id int) (*Project, bool, error) {
    query := postgres.SELECT(
        table.Projects.AllColumns,
    ).FROM(
        table.Projects,
    ).WHERE(
        table.Projects.ID.EQ(postgres.Int(int64(id))),
    )

    ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
    defer cancel()

    var dest model.Projects
    err := query.QueryContext(ctx, m.STDB, &dest)
    if err != nil {
        if errors.Is(err, qrm.ErrNoRows) {
            return nil, false, nil
        }
        return nil, false, err
    }

    return projectFromGen(dest), true, nil
}

func (m ProjectModel) GetByUUID(uuidStr string) (*Project, bool, error) {
    parsedUUID, err := uuid.Parse(uuidStr)
    if err != nil {
        return nil, false, err
    }

    query := postgres.SELECT(
        table.Projects.AllColumns,
    ).FROM(
        table.Projects,
    ).WHERE(
        table.Projects.UUID.EQ(postgres.UUID(parsedUUID)),
    )

    ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
    defer cancel()

    var dest model.Projects
    err = query.QueryContext(ctx, m.STDB, &dest)
    if err != nil {
        if errors.Is(err, qrm.ErrNoRows) {
            return nil, false, nil
        }
        return nil, false, err
    }

    return projectFromGen(dest), true, nil
}

func (m ProjectModel) GetAll() ([]*Project, error) {
    query := postgres.SELECT(
        table.Projects.AllColumns,
    ).FROM(
        table.Projects,
    )

    ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
    defer cancel()

    var dest []model.Projects
    err := query.QueryContext(ctx, m.STDB, &dest)
    if err != nil {
        return nil, err
    }

    projects := make([]*Project, len(dest))
    for i, d := range dest {
        projects[i] = projectFromGen(d)
    }

    return projects, nil
}

func (m ProjectModel) Update(project *Project) error {
    gen, err := projectToGen(project)
    if err != nil {
        return err
    }

    query := table.Projects.UPDATE(
        table.Projects.Name,
        table.Projects.Status,
        table.Projects.OrderedAt,
        table.Projects.OrderedBy,
        table.Projects.Version,
    ).MODEL(gen).WHERE(
        postgres.AND(
            table.Projects.ID.EQ(postgres.Int(int64(project.ID))),
            table.Projects.Version.EQ(postgres.Int(int64(project.Version))),
        ),
    ).RETURNING(
        table.Projects.UpdatedAt,
        table.Projects.Version,
    )

    ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
    defer cancel()

    var dest model.Projects
    err = query.QueryContext(ctx, m.STDB, &dest)
    if err != nil {
        return err
    }

    project.UpdatedAt = dest.UpdatedAt
    project.Version = int(dest.Version)

    return nil
}

func (m ProjectModel) Delete(id int) error {
    query := table.Projects.DELETE().WHERE(
        table.Projects.ID.EQ(postgres.Int(int64(id))),
    )

    ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
    defer cancel()

    _, err := query.ExecContext(ctx, m.STDB)
    return err
}
```

### Models Registry

Register all models in `models.go`:

```go
// models.go
package data

import (
    "database/sql"
    "github.com/jackc/pgx/v5/pgxpool"
)

type Models struct {
    DealershipAccounts DealershipAccountModel
    DealershipTokens   DealershipTokenModel
    DealershipUsers    DealershipUserModel
    Dealerships        DealershipModel
    InternalAccounts   InternalAccountModel
    InternalTokens     InternalTokenModel
    InternalUsers      InternalUserModel
    CatalogItems       CatalogItemModel
    Inlays             InlayModel
    InlayChats         InlayChatModel
    InlayProofs        InlayProofModel
    InlayMilestones    InlayMilestoneModel
    InlayBlockers      InlayBlockerModel
    Projects           ProjectModel
    ProjectChats       ProjectChatModel
    OrderSnapshots     OrderSnapshotModel
    Invoices           InvoiceModel
    Notifications      NotificationModel
    PriceGroups        PriceGroupModel
    Pool               *pgxpool.Pool
    STDB               *sql.DB
}

func NewModels(db *pgxpool.Pool, stdb *sql.DB) Models {
    return Models{
        DealershipAccounts: DealershipAccountModel{DB: db, STDB: stdb},
        DealershipTokens:   DealershipTokenModel{DB: db, STDB: stdb},
        DealershipUsers:    DealershipUserModel{DB: db, STDB: stdb},
        Dealerships:        DealershipModel{DB: db, STDB: stdb},
        // ... etc
        Pool: db,
        STDB: stdb,
    }
}
```

### Naming Conventions

| Concept    | TypeScript   | Go                  | SQL              |
| ---------- | ------------ | ------------------- | ---------------- |
| Fields     | snake_case   | CamelCase           | snake_case       |
| Types      | PascalCase   | PascalCase          | snake_case       |
| Enums      | string union | type + const struct | CHECK constraint |
| FK columns | `{table}_id` | `{Table}ID`         | `{table}_id`     |

### Regenerating Jet Code

After schema changes:

```bash
pnpm db:gen
# or
jet -dsn="postgresql://user:pass@localhost:5432/glassact?sslmode=disable" -path=./libs/data/pkg/gen
```

### Type Synchronization Checklist

When adding a new entity:

1. [ ] Add SQL table in migration
2. [ ] Add triggers in migration
3. [ ] Regenerate Jet code
4. [ ] Create TypeScript type in `libs/data/src/`
5. [ ] Create Go model in `libs/data/pkg/`
6. [ ] Register model in `models.go`
7. [ ] Export from `libs/data/src/index.ts`
