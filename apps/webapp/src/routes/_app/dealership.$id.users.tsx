import { createFileRoute } from "@tanstack/solid-router";
import { createSignal, For, Show } from "solid-js";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
  Form,
  Table,
  TableHeader,
  TableHead,
  TableRow,
  TableBody,
  TableCell,
  Button,
  TextField,
  TextFieldRoot,
  showToast,
} from "@glassact/ui";
import {
  createSolidTable,
  flexRender,
  getCoreRowModel,
  ColumnDef,
  getPaginationRowModel,
  getFilteredRowModel,
} from "@tanstack/solid-table";
import {
  GET,
  DealershipUser,
  DealershipUserRole,
  PERMISSION_ACTIONS,
} from "@glassact/data";
import { createForm } from "@tanstack/solid-form";
import { z } from "zod";
import { useMutation, useQuery, useQueryClient } from "@tanstack/solid-query";
import { getDealershipOpts } from "../../queries/dealership";
import {
  getDealershipUsersOpts,
  postDealershipUserOpts,
  patchDealershipUserOpts,
  deleteDealershipUserOpts,
} from "../../queries/user";
import { Can } from "../../components/Can";
import { isApiError } from "../../utils/is-api-error";

export const Route = createFileRoute("/_app/dealership/$id/users")({
  component: RouteComponent,
});

const roleOptions = [
  { label: "Viewer", value: "viewer" },
  { label: "Submitter", value: "submitter" },
  { label: "Approver", value: "approver" },
  { label: "Admin", value: "admin" },
];

const colors = [
  "FFB3BA",
  "FFDFBA",
  "FFFFBA",
  "BAFFC9",
  "BAE1FF",
  "E1BAFF",
  "F0BAFF",
  "BAFFEF",
  "FFD4BA",
];

function RouteComponent() {
  const params = Route.useParams();
  const queryClient = useQueryClient();

  const dealershipQuery = useQuery(() => getDealershipOpts(params().id));
  const dealershipId = () => dealershipQuery.data?.id;

  const usersQuery = useQuery(() => ({
    ...getDealershipUsersOpts(dealershipId()),
    enabled: dealershipId() !== undefined,
  }));

  const postUser = useMutation(() => postDealershipUserOpts());
  const deleteUser = useMutation(() => deleteDealershipUserOpts());

  const [addOpen, setAddOpen] = createSignal(false);
  const [editingUser, setEditingUser] =
    createSignal<GET<DealershipUser> | null>(null);

  function invalidateUsers() {
    queryClient.invalidateQueries({ queryKey: ["dealership-user"] });
  }

  function handleDeactivate(user: GET<DealershipUser>) {
    if (!confirm(`Deactivate ${user.name}? They will lose access.`)) return;
    deleteUser.mutate(user.uuid, {
      onSuccess() {
        showToast({
          title: "User deactivated",
          description: `${user.name} can no longer sign in.`,
          variant: "success",
        });
      },
      onError(error) {
        if (isApiError(error)) {
          showToast({
            title: "Problem deactivating user...",
            description: error?.data?.error ?? "Unknown error",
            variant: "error",
          });
        }
      },
      onSettled: invalidateUsers,
    });
  }

  const columns: ColumnDef<GET<DealershipUser>>[] = [
    { accessorKey: "name", header: "Name", cell: (info) => info.getValue() },
    { accessorKey: "email", header: "Email", cell: (info) => info.getValue() },
    { accessorKey: "role", header: "Role", cell: (info) => info.getValue() },
    {
      accessorKey: "is_active",
      header: "Status",
      cell: (info) => (info.getValue() ? "Active" : "Inactive"),
    },
    {
      id: "actions",
      header: "Actions",
      cell: (props) => (
        <Can permission={PERMISSION_ACTIONS.MANAGE_DEALERSHIP_USERS}>
          <div class="flex gap-2">
            <Button
              variant="outline"
              size="sm"
              onClick={() => setEditingUser(props.row.original)}
            >
              Edit
            </Button>
            <Show when={props.row.original.is_active}>
              <Button
                variant="ghost"
                size="sm"
                onClick={() => handleDeactivate(props.row.original)}
                disabled={deleteUser.isPending}
              >
                Deactivate
              </Button>
            </Show>
          </div>
        </Can>
      ),
    },
  ];

  const table = createSolidTable({
    get data() {
      return usersQuery.data ?? [];
    },
    columns,
    getCoreRowModel: getCoreRowModel(),
    getPaginationRowModel: getPaginationRowModel(),
    getFilteredRowModel: getFilteredRowModel(),
  });

  const addForm = createForm(() => ({
    defaultValues: {
      name: "",
      email: "",
      role: "" as DealershipUserRole,
    },
    validators: {
      onSubmit: z.object({
        name: z.string().min(1),
        email: z.email(),
        role: z.enum(["viewer", "submitter", "approver", "admin"]),
      }),
    },
    onSubmit: async ({ value }) => {
      const id = dealershipId();
      if (id === undefined) return;
      const color = colors[Math.floor(Math.random() * colors.length)];
      const body = {
        ...value,
        dealership_id: id,
        is_active: true,
        avatar: `https://ui-avatars.com/api/?name=${encodeURIComponent(value.name)}&background=${color}`,
      };
      postUser.mutate(body, {
        onSuccess() {
          setAddOpen(false);
          showToast({
            title: "Created new user!",
            description: `${value.name}'s account was created.`,
            variant: "success",
          });
          setTimeout(() => addForm.reset(), 300);
        },
        onError(error) {
          if (isApiError(error)) {
            showToast({
              title: "Problem creating new user...",
              description: error?.data?.error ?? "Unknown error",
              variant: "error",
            });
          }
        },
        onSettled: invalidateUsers,
      });
    },
  }));

  return (
    <div>
      <div class="mb-6">
        <h2 class="text-xl font-semibold">Users</h2>
        <p class="text-gray-600">Manage this dealership's team members.</p>
      </div>

      <div class="flex items-center justify-between py-4">
        <TextFieldRoot
          value={(table.getColumn("name")?.getFilterValue() as string) ?? ""}
          onChange={(value) => table.getColumn("name")?.setFilterValue(value)}
        >
          <TextField placeholder="Filter by name..." class="max-w-sm" />
        </TextFieldRoot>

        <Can permission={PERMISSION_ACTIONS.MANAGE_DEALERSHIP_USERS}>
          <Dialog open={addOpen()} onOpenChange={setAddOpen}>
            <DialogTrigger>
              <Button>Add a new user</Button>
            </DialogTrigger>
            <DialogContent>
              <DialogHeader>
                <DialogTitle>Add a new user</DialogTitle>
              </DialogHeader>
              <form
                onSubmit={(e) => {
                  e.preventDefault();
                  e.stopPropagation();
                  addForm.handleSubmit();
                }}
                class="flex flex-col gap-4"
              >
                <addForm.Field
                  name="name"
                  children={(field) => (
                    <Form.TextField field={field} label="Name" />
                  )}
                />
                <addForm.Field
                  name="email"
                  children={(field) => (
                    <Form.TextField field={field} label="Email" />
                  )}
                />
                <addForm.Field
                  name="role"
                  children={(field) => (
                    <Form.Combobox
                      field={field}
                      label="Role"
                      options={roleOptions}
                    />
                  )}
                />
                <Button type="submit" disabled={postUser.isPending}>
                  Add
                </Button>
              </form>
            </DialogContent>
          </Dialog>
        </Can>
      </div>

      <div class="rounded-md border">
        <Table>
          <TableHeader>
            <For each={table.getHeaderGroups()}>
              {(headerGroup) => (
                <TableRow>
                  <For each={headerGroup.headers}>
                    {(header) => (
                      <TableHead>
                        {header.isPlaceholder
                          ? null
                          : flexRender(
                              header.column.columnDef.header,
                              header.getContext(),
                            )}
                      </TableHead>
                    )}
                  </For>
                </TableRow>
              )}
            </For>
          </TableHeader>
          <TableBody>
            <Show
              when={table.getRowModel().rows?.length}
              fallback={
                <TableRow>
                  <TableCell colSpan={columns.length} class="h-24 text-center">
                    No results.
                  </TableCell>
                </TableRow>
              }
            >
              <For each={table.getRowModel().rows}>
                {(row) => (
                  <TableRow>
                    <For each={row.getVisibleCells()}>
                      {(cell) => (
                        <TableCell>
                          {flexRender(
                            cell.column.columnDef.cell,
                            cell.getContext(),
                          )}
                        </TableCell>
                      )}
                    </For>
                  </TableRow>
                )}
              </For>
            </Show>
          </TableBody>
        </Table>
      </div>

      <div class="flex items-center justify-end space-x-2 py-4">
        <Button
          variant="outline"
          size="sm"
          onClick={() => table.previousPage()}
          disabled={!table.getCanPreviousPage()}
        >
          Previous
        </Button>
        <Button
          variant="outline"
          size="sm"
          onClick={() => table.nextPage()}
          disabled={!table.getCanNextPage()}
        >
          Next
        </Button>
      </div>

      <Show when={editingUser()}>
        {(user) => (
          <EditUserDialog
            user={user()}
            onClose={() => setEditingUser(null)}
            onSaved={() => {
              setEditingUser(null);
              invalidateUsers();
            }}
          />
        )}
      </Show>
    </div>
  );
}

interface EditUserDialogProps {
  user: GET<DealershipUser>;
  onClose: () => void;
  onSaved: () => void;
}

function EditUserDialog(props: EditUserDialogProps) {
  const patchUser = useMutation(() => patchDealershipUserOpts());

  const form = createForm(() => ({
    defaultValues: {
      role: props.user.role as DealershipUserRole,
    },
    validators: {
      onSubmit: z.object({
        role: z.enum(["viewer", "submitter", "approver", "admin"]),
      }),
    },
    onSubmit: async ({ value }) => {
      patchUser.mutate(
        { uuid: props.user.uuid, body: { role: value.role } },
        {
          onSuccess() {
            showToast({
              title: "User updated",
              description: `${props.user.name}'s role was updated.`,
              variant: "success",
            });
            props.onSaved();
          },
          onError(error) {
            if (isApiError(error)) {
              showToast({
                title: "Problem updating user...",
                description: error?.data?.error ?? "Unknown error",
                variant: "error",
              });
            }
          },
        },
      );
    },
  }));

  return (
    <Dialog open onOpenChange={(open) => !open && props.onClose()}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Edit {props.user.name}</DialogTitle>
        </DialogHeader>
        <form
          onSubmit={(e) => {
            e.preventDefault();
            e.stopPropagation();
            form.handleSubmit();
          }}
          class="flex flex-col gap-4"
        >
          <form.Field
            name="role"
            children={(field) => (
              <Form.Combobox field={field} label="Role" options={roleOptions} />
            )}
          />
          <Button type="submit" disabled={patchUser.isPending}>
            Save
          </Button>
        </form>
      </DialogContent>
    </Dialog>
  );
}
