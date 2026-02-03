# GlassAct Studios - General Coding Rules

These rules apply to all code in this repository.

## Philosophy

### Correctness Over Cleverness

- Write code that is obviously correct, not code that is clever
- Prefer explicit over implicit
- If a solution requires explanation, consider simplifying it

### Dependencies Must Justify Their Weight

- Only add external dependencies when they provide substantial value
- For simple utilities, write them instead of importing a package
- Always check if the standard library solves the problem first
- Question any new dependency: "Is this worth the maintenance burden?"

**Approved high-value dependencies:**

- TanStack libraries (Query, Router, Form, Table)
- Kobalte (accessibility primitives)
- Zod (schema validation)
- Jet (SQL building)
- pgx (Postgres driver)
- validator/v10 (struct validation)

**Red flags - avoid:**

- `is-odd`, `left-pad` style micro-packages
- Packages that wrap standard library with minimal value
- Packages with many transitive dependencies for simple tasks

## Code Style

### No Redundant Comments

The code should be self-explanatory. Comments are only acceptable when explaining WHY, never WHAT.

```typescript
// BAD: Redundant comment
// Get the user by ID
const user = await getUserById(id);

// BAD: Explaining WHAT
// Loop through users and filter active ones
const activeUsers = users.filter((u) => u.isActive);

// GOOD: Explaining WHY (non-obvious business rule)
// Dealerships in trial period get 30-day invoice terms instead of 15
const invoiceTermDays = dealership.isInTrial ? 30 : 15;

// EVEN BETTER: Make the code more verbose instead of leaving a comment.
const trial30Day = 30;
const trial15Day = 15;
const invoiceTermDays = dealership.isInTrial ? trial30Day : trial15Day;

// GOOD: Explaining WHY (technical constraint)
// Using setTimeout because the DOM needs to settle before measuring
setTimeout(() => measureElement(), 0);
```

### Do Not Commit

- Committing is a human job
- Only suggest commit messages

### Naming Conventions

**Be descriptive over brief:**

```typescript
// BAD
const d = getDealership();
const u = users.filter((x) => x.a);

// GOOD
const dealership = getDealership();
const activeUsers = users.filter((user) => user.isActive);
```

**Booleans should read as questions:**

```typescript
// BAD
const active = true;
const blocker = hasHardBlocker();

// GOOD
const isActive = true;
const hasHardBlocker = checkForHardBlocker();
```

**Functions describe actions:**

```typescript
// BAD
function proof(inlayId: number) { ... }
function blocker(id: number) { ... }

// GOOD
function createProof(inlayId: number) { ... }
function resolveBlocker(id: number) { ... }
```

### File Organization

- One primary export per file
- Keep files focused on a single responsibility
- If a file exceeds 300 lines, consider splitting

### Error Handling

- Always handle errors explicitly
- Never swallow errors silently
- Provide meaningful error messages
- Log errors with context

```go
// BAD: Silent failure
result, _ := doSomething()

// BAD: Generic error
if err != nil {
    return errors.New("operation failed")
}

// GOOD: Contextual error
if err != nil {
    return fmt.Errorf("failed to create proof for inlay %d: %w", inlayID, err)
}
```

## Testing

### Test Behavior, Not Implementation

- Tests should verify outcomes, not internal mechanics
- If refactoring breaks tests without changing behavior, the tests are too coupled

### Integration Over Unit Where Practical

- Use testcontainers for database tests
- Test the full stack when feasible
- Unit test complex business logic

### Test Naming

```go
// BAD
func TestCreate(t *testing.T)
func TestProof1(t *testing.T)

// GOOD
func TestCreateProof_WithValidData_ReturnsProof(t *testing.T)
func TestCreateProof_WithMissingInlay_ReturnsError(t *testing.T)
```

## Git Practices

### Commit Messages

- First line: imperative mood, max 72 chars
- Body: explain WHY, not WHAT (the diff shows what)

```
Add proof approval workflow

Dealerships need to explicitly approve designs before ordering.
This prevents accidental orders of incorrect designs.
```

### Branch Naming

```
feature/proof-approval
fix/invoice-calculation
refactor/auth-middleware
```

## Security

### Never Commit Secrets

- Use environment variables
- Check `.env.example` exists for required vars
- Review diffs before committing

### Input Validation

- Validate all external input at API boundaries
- Use strong typing to prevent invalid states
- Sanitize data before storage

### Multi-Tenancy

- Every database query must be scoped to the appropriate dealership
- Never trust client-provided dealership IDs without verification
- Test permission boundaries explicitly
