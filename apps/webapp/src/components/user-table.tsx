import { createMemo, createSignal, For, JSX, Show } from "solid-js";
import {
  Button,
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  Form,
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
  TextField,
  TextFieldRoot,
  showToast,
} from "@glassact/ui";
import {
  ColumnDef,
  createSolidTable,
  flexRender,
  getCoreRowModel,
  getFilteredRowModel,
  getPaginationRowModel,
} from "@tanstack/solid-table";
import { createForm } from "@tanstack/solid-form";
import { z } from "zod";
import { Can } from "./Can";
import { isApiError } from "../utils/is-api-error";

export interface ManagedUser {
  uuid: string;
  name: string;
  email: string;
  role: string;
  is_active: boolean;
}

export interface RoleOption {
  label: string;
  value: string;
}

interface UserTableProps<T extends ManagedUser> {
  users: T[];
  permission: string;
  roleOptions: RoleOption[];
  /** Rendered in the toolbar, gated behind `permission` — typically an "Add a new user" dialog. */
  addAction?: JSX.Element;
  /** Inserted between Role and Status, for context a single view needs (e.g. which dealership). */
  extraColumns?: ColumnDef<T>[];
  onUpdateRole: (user: T, role: string) => Promise<unknown>;
  onSetActive: (user: T, isActive: boolean) => Promise<unknown>;
}

function toErrorMessage(error: unknown) {
  if (error instanceof Error && isApiError(error)) {
    return error.data?.error ?? "Unknown error";
  }
  return "Unknown error";
}

export function UserTable<T extends ManagedUser>(props: UserTableProps<T>) {
  const [editingUser, setEditingUser] = createSignal<T | null>(null);
  const [pendingUuid, setPendingUuid] = createSignal<string | null>(null);

  // Only deactivation is destructive enough to confirm — restoring access is not.
  async function handleSetActive(user: T, isActive: boolean) {
    if (
      !isActive &&
      !confirm(`Deactivate ${user.name}? They will lose access.`)
    ) {
      return;
    }

    setPendingUuid(user.uuid);
    try {
      await props.onSetActive(user, isActive);
      showToast({
        title: isActive ? "User reactivated" : "User deactivated",
        description: isActive
          ? `${user.name} can sign in again.`
          : `${user.name} can no longer sign in.`,
        variant: "success",
      });
    } catch (error) {
      showToast({
        title: isActive
          ? "Problem reactivating user..."
          : "Problem deactivating user...",
        description: toErrorMessage(error),
        variant: "error",
      });
    } finally {
      setPendingUuid(null);
    }
  }

  const columns = createMemo<ColumnDef<T>[]>(() => [
    { accessorKey: "name", header: "Name", cell: (info) => info.getValue() },
    { accessorKey: "email", header: "Email", cell: (info) => info.getValue() },
    { accessorKey: "role", header: "Role", cell: (info) => info.getValue() },
    ...(props.extraColumns ?? []),
    {
      accessorKey: "is_active",
      header: "Status",
      cell: (info) => (info.getValue() ? "Active" : "Inactive"),
    },
    {
      id: "actions",
      header: "Actions",
      cell: (info) => (
        <Can permission={props.permission}>
          <div class="flex gap-2">
            <Button
              variant="outline"
              size="sm"
              onClick={() => setEditingUser(() => info.row.original)}
            >
              Edit
            </Button>
            <Button
              variant="ghost"
              size="sm"
              onClick={() =>
                handleSetActive(
                  info.row.original,
                  !info.row.original.is_active,
                )
              }
              disabled={pendingUuid() === info.row.original.uuid}
            >
              {info.row.original.is_active ? "Deactivate" : "Reactivate"}
            </Button>
          </div>
        </Can>
      ),
    },
  ]);

  const table = createSolidTable({
    get data() {
      return props.users;
    },
    get columns() {
      return columns();
    },
    getCoreRowModel: getCoreRowModel(),
    getPaginationRowModel: getPaginationRowModel(),
    getFilteredRowModel: getFilteredRowModel(),
  });

  return (
    <div>
      <div class="flex items-center justify-between py-4">
        <TextFieldRoot
          value={(table.getColumn("name")?.getFilterValue() as string) ?? ""}
          onChange={(value) => table.getColumn("name")?.setFilterValue(value)}
        >
          <TextField placeholder="Filter by name..." class="max-w-sm" />
        </TextFieldRoot>

        <Can permission={props.permission}>{props.addAction}</Can>
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
                  <TableCell
                    colSpan={columns().length}
                    class="h-24 text-center"
                  >
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
            roleOptions={props.roleOptions}
            onSave={(role) => props.onUpdateRole(user(), role)}
            onClose={() => setEditingUser(null)}
          />
        )}
      </Show>
    </div>
  );
}

interface EditUserDialogProps<T extends ManagedUser> {
  user: T;
  roleOptions: RoleOption[];
  onSave: (role: string) => Promise<unknown>;
  onClose: () => void;
}

function EditUserDialog<T extends ManagedUser>(props: EditUserDialogProps<T>) {
  const [isSaving, setIsSaving] = createSignal(false);

  const form = createForm(() => ({
    defaultValues: {
      role: props.user.role,
    },
    validators: {
      onSubmit: z.object({
        role: z.string().min(1, "Select a role"),
      }),
    },
    onSubmit: async ({ value }) => {
      setIsSaving(true);
      try {
        await props.onSave(value.role);
        showToast({
          title: "User updated",
          description: `${props.user.name}'s role was updated.`,
          variant: "success",
        });
        props.onClose();
      } catch (error) {
        showToast({
          title: "Problem updating user...",
          description: toErrorMessage(error),
          variant: "error",
        });
      } finally {
        setIsSaving(false);
      }
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
              <Form.Combobox
                field={field}
                label="Role"
                options={props.roleOptions}
              />
            )}
          />
          <Button type="submit" disabled={isSaving()}>
            Save
          </Button>
        </form>
      </DialogContent>
    </Dialog>
  );
}
