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
import { GET, DealershipUser, DealershipUserRole } from "@glassact/data";
import { createForm } from "@tanstack/solid-form";
import { z } from "zod";
import { useMutation, useQuery, useQueryClient } from "@tanstack/solid-query";
import {
  getDealershipUsersOpts,
  postDealershipUserOpts,
} from "../../queries/user";
import DealershipCombobox from "../../components/dealership-combobox";
import { isApiError } from "../../utils/is-api-error";

export const Route = createFileRoute("/_app/admin/users/dealership")({
  component: RouteComponent,
});

const defaultColumns: ColumnDef<GET<DealershipUser>>[] = [
  {
    accessorKey: "name",
    header: "Name",
    cell: (info) => info.getValue(),
  },
  {
    accessorKey: "email",
    header: "Email",
    cell: (info) => info.getValue(),
  },
  {
    accessorKey: "role",
    header: "Role",
    cell: (info) => info.getValue(),
  },
  {
    accessorKey: "dealership_id",
    header: "Dealership ID",
    cell: (info) => info.getValue(),
  },
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
  const query = useQuery(getDealershipUsersOpts);
  const queryClient = useQueryClient();
  const postUser = useMutation(postDealershipUserOpts);

  const table = createSolidTable({
    get data() {
      return query.data ?? [];
    },
    columns: defaultColumns,
    getCoreRowModel: getCoreRowModel(),
    getPaginationRowModel: getPaginationRowModel(),
    getFilteredRowModel: getFilteredRowModel(),
  });

  const [dialogOpen, setDialogOpen] = createSignal(false);

  const form = createForm(() => ({
    defaultValues: {
      name: "",
      email: "",
      dealership_id: undefined as unknown as number,
      role: "" as DealershipUserRole,
    },
    validators: {
      onSubmit: z.object({
        name: z.string().min(1),
        email: z.email(),
        dealership_id: z.number().int(),
        role: z.enum(["viewer", "submitter", "approver", "admin"]),
      }),
    },
    onSubmit: async ({ value }) => {
      const color = colors[Math.floor(Math.random() * colors.length)];
      const body = {
        ...value,
        is_active: true,
        avatar: `https://ui-avatars.com/api/?name=${encodeURIComponent(value.name)}&background=${color}`,
      };
      postUser.mutate(body, {
        onSuccess() {
          setDialogOpen(false);
          showToast({
            title: "Created new user!",
            description: `${value.name}'s account was created.`,
            variant: "success",
          });
          setTimeout(() => {
            form.reset();
          }, 300);
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
        onSettled() {
          queryClient.invalidateQueries({
            queryKey: ["dealership-user"],
          });
        },
      });
    },
  }));

  return (
    <div>
      <div class="mb-6">
        <h1 class="text-2xl font-bold">Dealership Users</h1>
        <p class="text-gray-600">Manage users from dealerships</p>
      </div>

      <div class="flex items-center justify-between py-4">
        <TextFieldRoot
          value={(table.getColumn("name")?.getFilterValue() as string) ?? ""}
          onChange={(value) =>
            table.getColumn("name")?.setFilterValue(value)
          }
        >
          <TextField placeholder="Filter by name..." class="max-w-sm" />
        </TextFieldRoot>
        <Dialog open={dialogOpen()} onOpenChange={setDialogOpen}>
          <DialogTrigger>
            <Button>Add a new user</Button>
          </DialogTrigger>
          <DialogContent>
            <DialogHeader>
              <DialogTitle>Add a new dealership user</DialogTitle>
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
                name="name"
                children={(field) => (
                  <Form.TextField field={field} label="Name" />
                )}
              />

              <form.Field
                name="email"
                children={(field) => (
                  <Form.TextField field={field} label="Email" />
                )}
              />

              <form.Field
                name="dealership_id"
                children={(field) => <DealershipCombobox field={field} />}
              />

              <form.Field
                name="role"
                children={(field) => (
                  <Form.Combobox
                    field={field}
                    label="Role"
                    options={[
                      { label: "Viewer", value: "viewer" },
                      { label: "Submitter", value: "submitter" },
                      { label: "Approver", value: "approver" },
                      { label: "Admin", value: "admin" },
                    ]}
                  />
                )}
              />

              <Button type="submit" disabled={postUser.isPending}>
                Add
              </Button>
            </form>
          </DialogContent>
        </Dialog>
      </div>
      <div class="rounded-md border">
        <Table>
          <TableHeader>
            <For each={table.getHeaderGroups()}>
              {(headerGroup) => (
                <TableRow>
                  <For each={headerGroup.headers}>
                    {(header) => {
                      return (
                        <TableHead>
                          {header.isPlaceholder
                            ? null
                            : flexRender(
                                header.column.columnDef.header,
                                header.getContext(),
                              )}
                        </TableHead>
                      );
                    }}
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
                    colSpan={defaultColumns.length}
                    class="h-24 text-center"
                  >
                    No results.
                  </TableCell>
                </TableRow>
              }
            >
              <For each={table.getRowModel().rows}>
                {(row) => (
                  <TableRow data-state={row.getIsSelected() && "selected"}>
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
    </div>
  );
}
