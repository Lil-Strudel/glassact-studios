# GlassAct Studios - Frontend Rules

These rules apply to all frontend code in `apps/webapp/` and `libs/ui/`.

## Tech Stack

| Purpose      | Library         |
| ------------ | --------------- |
| UI Framework | SolidJS         |
| Routing      | TanStack Router |
| Server State | TanStack Query  |
| Forms        | TanStack Form   |
| Tables       | TanStack Table  |
| Styling      | Tailwind CSS    |
| Primitives   | Kobalte         |
| Validation   | Zod             |

## SolidJS-Specific Rules

### This is NOT React

SolidJS has different mental models. Do not apply React patterns blindly.

**Signals are called to get values:**

```tsx
// BAD: Treating signal like React state
const [count, setCount] = createSignal(0);
return <div>{count}</div>; // Won't be reactive!

// GOOD: Call the signal
return <div>{count()}</div>;
```

**Destructuring props breaks reactivity:**

```tsx
// BAD: Destructured props lose reactivity
function UserCard({ name, email }) {
  return <div>{name}</div>; // Won't update!
}

// GOOD: Access props directly
function UserCard(props) {
  return <div>{props.name}</div>;
}

// ALSO GOOD: Use splitProps for specific needs
function UserCard(props) {
  const [local, rest] = splitProps(props, ["name"]);
  return <div {...rest}>{local.name}</div>;
}
```

**Use createMemo for derived state:**

```tsx
// BAD: Recalculating on every render
function Component() {
  const items = useItems();
  const total = items().reduce((sum, i) => sum + i.price, 0);
  return <div>{total}</div>;
}

// GOOD: Memoized derivation
function Component() {
  const items = useItems();
  const total = createMemo(() => items().reduce((sum, i) => sum + i.price, 0));
  return <div>{total()}</div>;
}
```

**createEffect sparingly:**

```tsx
// BAD: Effect for derived state
createEffect(() => {
  setFullName(`${firstName()} ${lastName()}`);
});

// GOOD: Use createMemo instead
const fullName = createMemo(() => `${firstName()} ${lastName()}`);

// ACCEPTABLE: Effect for side effects
createEffect(() => {
  document.title = `Project: ${projectName()}`;
});
```

## Type Safety

### Use Types from @glassact/data

All API types come from the shared data library. Never define API types in the frontend.

```typescript
import type { GET, POST, Project, Inlay } from "@glassact/data";

// Response types
type ProjectResponse = GET<Project>;

// Request types
type CreateProjectRequest = POST<Project>;
```

### Never Use `any`

```typescript
// BAD
function processData(data: any) { ... }

// GOOD: Use unknown with narrowing
function processData(data: unknown) {
  if (isProject(data)) {
    // Now TypeScript knows it's a Project
  }
}

// GOOD: Use generics
function processData<T>(data: T): ProcessedData<T> { ... }
```

### Zod for Runtime Validation

```typescript
import { z } from "zod";

const ProjectSchema = z.object({
  name: z.string().min(1),
  dealership_id: z.number().positive(),
});

// Validate external data
const validated = ProjectSchema.parse(untrustedData);
```

## Query Patterns

All queries go in `apps/webapp/src/queries/`. Follow this pattern:

```typescript
// queries/project.ts
import { queryOptions } from "@tanstack/solid-query";
import api from "./api";
import type { GET, POST, Project } from "@glassact/data";
import { mutationOptions } from "../utils/mutation-options";

// 1. Raw fetch function
export async function getProject(uuid: string): Promise<GET<Project>> {
  const res = await api.get(`/project/${uuid}`);
  return res.data;
}

// 2. Query options factory (returns a function for reactivity)
export function getProjectOpts(uuid: string) {
  return queryOptions({
    queryKey: ["project", uuid],
    queryFn: () => getProject(uuid),
  });
}

// 3. For lists
export async function getProjects(): Promise<GET<Project>[]> {
  const res = await api.get("/project");
  return res.data;
}

export function getProjectsOpts() {
  return queryOptions({
    queryKey: ["project"],
    queryFn: getProjects,
  });
}

// 4. Mutations
export async function postProject(body: POST<Project>): Promise<GET<Project>> {
  const res = await api.post("/project", body);
  return res.data;
}

export function postProjectOpts() {
  return mutationOptions({
    mutationFn: postProject,
  });
}
```

**Usage in components:**

```tsx
function ProjectDetail(props: { uuid: string }) {
  const query = useQuery(() => getProjectOpts(props.uuid)());

  return (
    <Show when={query.data} fallback={<Loading />}>
      {(project) => <div>{project().name}</div>}
    </Show>
  );
}
```

**Query Key Conventions:**

```typescript
// Entity list
["project"]["inlay"][
  // Single entity
  ("project", uuid)
][("inlay", uuid)][
  // Nested resources
  ("project", projectUuid, "inlays")
][("inlay", inlayUuid, "proofs")][
  // Filtered lists
  ("project", { status: "ordered" })
];
```

## Permission Handling

### The `<Can>` Component

All permission checks should use the centralized `<Can>` component. This allows permission logic to be updated in one place.

```tsx
// GOOD: Using <Can>
<Can permission="approve_designs">
  <Button onClick={handleApprove}>Approve</Button>
</Can>;

// BAD: Inline permission checks
{
  user().role === "approver" && (
    <Button onClick={handleApprove}>Approve</Button>
  );
}
```

### Permission Hook

For programmatic checks:

```tsx
function useCanApprove() {
  const { hasPermission } = usePermissions();
  return hasPermission("approve_designs");
}

function ProofActions() {
  const canApprove = useCanApprove();

  const handleSubmit = () => {
    if (!canApprove) {
      toast.error("You don't have permission to approve");
      return;
    }
    // ...
  };
}
```

## Component Patterns

### One Component Per File

```
components/
├── inlay-card.tsx      # InlayCard
├── proof-viewer.tsx    # ProofViewer
└── chat-message.tsx    # ChatMessage
```

### Props Interface Pattern

```tsx
interface InlayCardProps {
  inlay: GET<Inlay>;
  onSelect?: (inlay: GET<Inlay>) => void;
  isSelected?: boolean;
}

export function InlayCard(props: InlayCardProps) {
  // Access props.inlay, props.onSelect, etc.
}
```

### Composition Over Configuration

```tsx
// BAD: Overly configurable
<Card
  title="Project"
  showActions={true}
  actionPosition="top-right"
  actionButtons={[{ label: "Edit", onClick: ... }]}
/>

// GOOD: Composable
<Card>
  <Card.Header>
    <Card.Title>Project</Card.Title>
    <Card.Actions>
      <Button onClick={...}>Edit</Button>
    </Card.Actions>
  </Card.Header>
  <Card.Content>...</Card.Content>
</Card>
```

## Forms with TanStack Form

```tsx
import { createForm } from "@tanstack/solid-form";
import { zodValidator } from "@tanstack/zod-form-adapter";

const CreateProjectSchema = z.object({
  name: z.string().min(1, "Name is required"),
  dealership_id: z.number().positive(),
});

function CreateProjectForm() {
  const form = createForm(() => ({
    defaultValues: {
      name: "",
      dealership_id: 0,
    },
    onSubmit: async ({ value }) => {
      await createProject(value);
    },
    validatorAdapter: zodValidator(),
  }));

  return (
    <form
      onSubmit={(e) => {
        e.preventDefault();
        form.handleSubmit();
      }}
    >
      <form.Field
        name="name"
        validators={{ onChange: CreateProjectSchema.shape.name }}
      >
        {(field) => (
          <div>
            <Input
              value={field().state.value}
              onInput={(e) => field().handleChange(e.target.value)}
            />
            <Show when={field().state.meta.errors.length > 0}>
              <span class="text-red-500">
                {field().state.meta.errors.join(", ")}
              </span>
            </Show>
          </div>
        )}
      </form.Field>

      <Button type="submit" disabled={form.state.isSubmitting}>
        {form.state.isSubmitting ? "Creating..." : "Create"}
      </Button>
    </form>
  );
}
```

## Styling with Tailwind

### Use Design System Classes

Prefer semantic/design system classes from the UI library over raw Tailwind:

```tsx
// Prefer UI library components
<Button variant="primary" size="sm">Save</Button>

// Over raw Tailwind
<button class="bg-blue-500 text-white px-4 py-2 rounded">Save</button>
```

### Consistent Spacing

Use Tailwind's spacing scale consistently:

- `gap-2`, `gap-4`, `gap-6` for flex/grid gaps
- `p-4`, `p-6` for card padding
- `mb-4`, `mt-6` for vertical rhythm

### Responsive Design

Mobile-first approach:

```tsx
<div class="flex flex-col md:flex-row gap-4">
  <div class="w-full md:w-1/3">Sidebar</div>
  <div class="w-full md:w-2/3">Content</div>
</div>
```

## Error Handling

### Query Error States

```tsx
function ProjectList() {
  const query = useQuery(() => getProjectsOpts()());

  return (
    <Switch>
      <Match when={query.isLoading}>
        <LoadingSpinner />
      </Match>
      <Match when={query.isError}>
        <ErrorMessage error={query.error} retry={query.refetch} />
      </Match>
      <Match when={query.data}>
        {(projects) => (
          <For each={projects()}>
            {(project) => <ProjectCard project={project} />}
          </For>
        )}
      </Match>
    </Switch>
  );
}
```

### Mutation Error Handling

```tsx
const mutation = createMutation(() => postProjectOpts());

const handleSubmit = async () => {
  try {
    await mutation.mutateAsync(formData);
    toast.success("Project created");
    navigate({ to: "/projects" });
  } catch (error) {
    if (isApiError(error)) {
      toast.error(error.message);
    } else {
      toast.error("An unexpected error occurred");
    }
  }
};
```

## Route Organization

```
routes/
├── __root.tsx           # Root layout
├── _app.tsx             # Authenticated layout
├── _app/
│   ├── dashboard.tsx    # /dashboard
│   ├── projects.tsx     # /projects
│   ├── projects_.$id.tsx # /projects/:id
│   └── admin/
│       ├── users.tsx    # /admin/users
│       └── catalog.tsx  # /admin/catalog
├── login.tsx            # /login
└── index.tsx            # / (redirect)
```

### Route Data Loading

```tsx
export const Route = createFileRoute("/_app/projects/$id")({
  loader: ({ params }) => ({
    project: getProject(params.id),
  }),
  component: ProjectDetail,
});

function ProjectDetail() {
  const { project } = Route.useLoaderData();
  // project is already loaded
}
```
