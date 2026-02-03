# GlassAct Studios - Backend Rules

These rules apply to all Go code in `apps/api/` and `libs/data/pkg/`.

## Tech Stack

| Purpose | Library |
|---------|---------|
| HTTP | Standard library `net/http` |
| Routing | Standard library `http.ServeMux` |
| Middleware | `justinas/alice` |
| Database | `pgx/v5` (driver), `go-jet/jet` (query builder) |
| Validation | `go-playground/validator/v10` |
| UUID | `google/uuid` |
| Testing | `testify`, `testcontainers-go` |

## Project Structure

```
apps/api/
├── cmd/api/
│   └── main.go              # Entry point
├── app/
│   ├── app.go               # Application struct
│   ├── context.go           # Request context helpers
│   ├── errors.go            # Error definitions
│   ├── handlers.go          # Common handlers
│   ├── helpers.go           # JSON helpers
│   ├── middleware.go        # Auth, logging, etc.
│   └── utils.go             # Utilities
├── config/
│   └── config.go            # Configuration
└── modules/
    ├── modules.go           # Route registration
    ├── auth/
    │   ├── authHandlers.go
    │   └── authServices.go
    ├── project/
    │   ├── projectHandlers.go
    │   └── projectServices.go
    └── ...
```

## Module Pattern

Each feature gets its own module with handlers and services:

```go
// modules/project/projectHandlers.go
package project

import (
    "net/http"
    "github.com/Lil-Strudel/glassact-studios/apps/api/app"
    "github.com/Lil-Strudel/glassact-studios/libs/data/pkg"
)

type ProjectModule struct {
    *app.Application
}

func NewProjectModule(app *app.Application) *ProjectModule {
    return &ProjectModule{app}
}

func (m *ProjectModule) HandleGetProjects(w http.ResponseWriter, r *http.Request) {
    // Handler implementation
}
```

**Register routes in modules.go:**
```go
func GetRoutes(app *app.Application) http.Handler {
    mux := http.NewServeMux()
    
    protected := alice.New(app.Authenticate)
    
    projectModule := project.NewProjectModule(app)
    mux.Handle("GET /api/project", protected.ThenFunc(projectModule.HandleGetProjects))
    mux.Handle("GET /api/project/{uuid}", protected.ThenFunc(projectModule.HandleGetProjectByUUID))
    mux.Handle("POST /api/project", protected.ThenFunc(projectModule.HandlePostProject))
    
    return mux
}
```

## Handler Pattern

```go
func (m *ProjectModule) HandlePostProject(w http.ResponseWriter, r *http.Request) {
    // 1. Define request body struct with validation tags
    var body struct {
        Name         string `json:"name" validate:"required"`
        DealershipID int    `json:"dealership_id" validate:"required,gt=0"`
    }

    // 2. Parse and validate request
    err := m.ReadJSONBody(w, r, &body)
    if err != nil {
        m.WriteError(w, r, m.Err.BadRequest, err)
        return
    }

    // 3. Get authenticated user for authorization
    user := m.ContextGetUser(r)
    
    // 4. Authorization check (if needed beyond middleware)
    if user.DealershipID != body.DealershipID {
        m.WriteError(w, r, m.Err.Forbidden, nil)
        return
    }

    // 5. Business logic
    project := data.Project{
        Name:         body.Name,
        DealershipID: body.DealershipID,
        Status:       data.ProjectStatuses.Draft,
    }

    err = m.Db.Projects.Insert(&project)
    if err != nil {
        m.WriteError(w, r, m.Err.ServerError, err)
        return
    }

    // 6. Return response
    m.WriteJSON(w, r, http.StatusCreated, project)
}
```

## Data Layer Pattern

All database models live in `libs/data/pkg/`:

```go
// libs/data/pkg/projects.go
package data

import (
    "context"
    "database/sql"
    "time"
    
    "github.com/Lil-Strudel/glassact-studios/libs/data/pkg/gen/glassact/public/model"
    "github.com/Lil-Strudel/glassact-studios/libs/data/pkg/gen/glassact/public/table"
    "github.com/go-jet/jet/v2/postgres"
)

// 1. Define the API model
type Project struct {
    StandardTable
    Name         string        `json:"name"`
    Status       ProjectStatus `json:"status"`
    DealershipID int           `json:"dealership_id"`
}

// 2. Define the model struct
type ProjectModel struct {
    DB   *pgxpool.Pool
    STDB *sql.DB
}

// 3. Conversion from Jet-generated model
func projectFromGen(gen model.Projects) *Project {
    return &Project{
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
}

// 4. Conversion to Jet-generated model
func projectToGen(p *Project) (*model.Projects, error) {
    var projectUUID uuid.UUID
    var err error
    
    if p.UUID != "" {
        projectUUID, err = uuid.Parse(p.UUID)
        if err != nil {
            return nil, err
        }
    }
    
    return &model.Projects{
        ID:           int32(p.ID),
        UUID:         projectUUID,
        Name:         p.Name,
        Status:       string(p.Status),
        DealershipID: int32(p.DealershipID),
        CreatedAt:    p.CreatedAt,
        UpdatedAt:    p.UpdatedAt,
        Version:      int32(p.Version),
    }, nil
}

// 5. CRUD operations
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

    // Update the passed-in struct with generated values
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

    // ... same pattern as GetByID
}

func (m ProjectModel) Update(project *Project) error {
    gen, err := projectToGen(project)
    if err != nil {
        return err
    }

    query := table.Projects.UPDATE(
        table.Projects.Name,
        table.Projects.Status,
        table.Projects.Version, // Include version for optimistic locking
    ).MODEL(gen).WHERE(
        postgres.AND(
            table.Projects.ID.EQ(postgres.Int(int64(project.ID))),
            table.Projects.Version.EQ(postgres.Int(int64(project.Version))),
        ),
    ).RETURNING(
        table.Projects.UpdatedAt,
        table.Projects.Version,
    )

    // ... execute and update project struct
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

## Transaction Pattern

For operations that span multiple tables:

```go
func (m *ProjectModule) HandlePostProjectWithInlays(w http.ResponseWriter, r *http.Request) {
    var body struct {
        Name   string `json:"name" validate:"required"`
        Inlays []struct {
            Name string `json:"name" validate:"required"`
            Type string `json:"type" validate:"required"`
        } `json:"inlays" validate:"required,dive"`
    }

    err := m.ReadJSONBody(w, r, &body)
    if err != nil {
        m.WriteError(w, r, m.Err.BadRequest, err)
        return
    }

    // Start transaction
    tx, err := m.Db.STDB.Begin()
    if err != nil {
        m.WriteError(w, r, m.Err.ServerError, err)
        return
    }
    defer tx.Rollback() // Rollback if not committed

    // Create project
    project := data.Project{Name: body.Name, ...}
    err = m.Db.Projects.TxInsert(tx, &project)
    if err != nil {
        m.WriteError(w, r, m.Err.ServerError, err)
        return
    }

    // Create inlays
    for _, inlayBody := range body.Inlays {
        inlay := data.Inlay{
            ProjectID: project.ID,
            Name:      inlayBody.Name,
            Type:      data.InlayType(inlayBody.Type),
        }
        err = m.Db.Inlays.TxInsert(tx, &inlay)
        if err != nil {
            m.WriteError(w, r, m.Err.ServerError, err)
            return
        }
    }

    // Commit transaction
    err = tx.Commit()
    if err != nil {
        m.WriteError(w, r, m.Err.ServerError, err)
        return
    }

    m.WriteJSON(w, r, http.StatusCreated, project)
}
```

## Authentication Middleware

```go
func (app *Application) Authenticate(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Header().Add("Vary", "Authorization")

        authHeader := r.Header.Get("Authorization")
        if authHeader == "" {
            app.WriteError(w, r, app.Err.AuthenticationError, nil)
            return
        }

        parts := strings.Split(authHeader, " ")
        if len(parts) != 2 || parts[0] != "Bearer" {
            app.WriteError(w, r, app.Err.AuthenticationError, nil)
            return
        }

        token := parts[1]

        // Try dealership user first
        user, found, err := app.Db.DealershipUsers.GetForToken(data.ScopeAccess, token)
        if err != nil {
            app.WriteError(w, r, app.Err.ServerError, err)
            return
        }
        
        if found {
            r = app.ContextSetDealershipUser(r, user)
            next.ServeHTTP(w, r)
            return
        }

        // Try internal user
        internalUser, found, err := app.Db.InternalUsers.GetForToken(data.ScopeAccess, token)
        if err != nil {
            app.WriteError(w, r, app.Err.ServerError, err)
            return
        }

        if found {
            r = app.ContextSetInternalUser(r, internalUser)
            next.ServeHTTP(w, r)
            return
        }

        app.WriteError(w, r, app.Err.AuthenticationError, nil)
    })
}
```

## Error Handling

### Defined Errors
```go
// app/errors.go
type appError struct {
    ServerError        AppError
    BadRequest         AppError
    AuthenticationError AppError
    Forbidden          AppError
    RecordNotFound     AppError
    // ...
}

type AppError struct {
    StatusCode int
    Message    string
}
```

### Writing Errors
```go
func (app *Application) WriteError(w http.ResponseWriter, r *http.Request, appErr AppError, err error) {
    if err != nil {
        app.Log.Error(appErr.Message, "error", err, "path", r.URL.Path)
    }

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(appErr.StatusCode)
    
    json.NewEncoder(w).Encode(map[string]string{
        "error": appErr.Message,
    })
}
```

### Error Wrapping
```go
// BAD: Lost context
if err != nil {
    return err
}

// GOOD: Add context
if err != nil {
    return fmt.Errorf("failed to create proof for inlay %d: %w", inlayID, err)
}
```

## Validation

### Struct Tags
```go
var body struct {
    Name         string  `json:"name" validate:"required,min=1,max=255"`
    Email        string  `json:"email" validate:"required,email"`
    DealershipID int     `json:"dealership_id" validate:"required,gt=0"`
    Status       string  `json:"status" validate:"required,oneof=draft active"`
    Price        float64 `json:"price" validate:"gte=0"`
    Tags         []string `json:"tags" validate:"dive,min=1,max=50"`
}
```

### Custom Validation
```go
// For complex validation beyond struct tags
func validateOrderPlacement(project *data.Project, inlays []*data.Inlay) error {
    for _, inlay := range inlays {
        if inlay.ApprovedProofID == nil {
            return fmt.Errorf("inlay %s has no approved proof", inlay.Name)
        }
    }
    return nil
}
```

## Testing

### Integration Tests with Testcontainers
```go
func TestProjectModel_Insert(t *testing.T) {
    ctx := context.Background()
    
    // Start Postgres container
    container, err := postgres.Run(ctx,
        "postgres:16-alpine",
        postgres.WithDatabase("test"),
        postgres.WithUsername("test"),
        postgres.WithPassword("test"),
    )
    require.NoError(t, err)
    defer container.Terminate(ctx)

    // Run migrations
    connStr, err := container.ConnectionString(ctx)
    require.NoError(t, err)
    
    // ... run migrations ...

    // Create model
    pool, err := pgxpool.New(ctx, connStr)
    require.NoError(t, err)
    
    model := data.ProjectModel{DB: pool, STDB: sqlDB}

    // Test
    project := &data.Project{
        Name:         "Test Project",
        Status:       data.ProjectStatuses.Draft,
        DealershipID: 1,
    }
    
    err = model.Insert(project)
    require.NoError(t, err)
    assert.NotZero(t, project.ID)
    assert.NotEmpty(t, project.UUID)
}
```

### Table-Driven Tests
```go
func TestProjectStatus_Transitions(t *testing.T) {
    tests := []struct {
        name        string
        from        data.ProjectStatus
        to          data.ProjectStatus
        shouldError bool
    }{
        {"draft to designing", data.ProjectStatuses.Draft, data.ProjectStatuses.Designing, false},
        {"designing to ordered", data.ProjectStatuses.Designing, data.ProjectStatuses.Ordered, true},
        {"approved to ordered", data.ProjectStatuses.Approved, data.ProjectStatuses.Ordered, false},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := validateStatusTransition(tt.from, tt.to)
            if tt.shouldError {
                assert.Error(t, err)
            } else {
                assert.NoError(t, err)
            }
        })
    }
}
```

## Logging

```go
// Use structured logging
app.Log.Info("project created",
    "project_id", project.ID,
    "dealership_id", project.DealershipID,
    "user_id", user.ID,
)

app.Log.Error("failed to create invoice",
    "error", err,
    "project_id", projectID,
)

// Don't log sensitive data
// BAD
app.Log.Info("user login", "email", email, "token", token)

// GOOD
app.Log.Info("user login", "email", email, "user_id", user.ID)
```
