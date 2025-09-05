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
import { GET, User } from "@glassact/data";
import { createForm } from "@tanstack/solid-form";
import { z } from "zod";
import { useMutation, useQuery, useQueryClient } from "@tanstack/solid-query";
import { getUsersOpts, postUserOpts } from "../../queries/user";
import DealershipCombobox from "../../components/dealership-combobox";
import { isApiError } from "../../utils/is-api-error";

export const Route = createFileRoute("/_app/admin/users")({
  component: RouteComponent,
});

const defaultColumns: ColumnDef<GET<User>>[] = [
  {
    accessorKey: "name",
    header: "Name",
    cell: (info) => info.getValue(),
    footer: (info) => info.column.id,
  },
  {
    accessorKey: "email",
    header: "Email",
    cell: (info) => info.getValue(),
    footer: (info) => info.column.id,
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
  const query = useQuery(getUsersOpts);
  const queryClient = useQueryClient();
  const postUser = useMutation(postUserOpts);

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
    },
    validators: {
      onSubmit: z.object({
        name: z.string().min(1),
        email: z.email(),
        dealership_id: z.number().int(),
      }),
    },
    onSubmit: async ({ value }) => {
      const color = colors[Math.floor(Math.random() * colors.length)];
      const body = {
        ...value,
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
          queryClient.invalidateQueries({ queryKey: ["user"] });
        },
      });
    },
  }));

  return (
    <div>
      <div class="flex items-center justify-between py-4">
        <TextFieldRoot
          value={(table.getColumn("name")?.getFilterValue() as string) ?? ""}
          onChange={(value) => table.getColumn("name")?.setFilterValue(value)}
        >
          <TextField placeholder="Filter by name..." class="max-w-sm" />
        </TextFieldRoot>
        <Dialog open={dialogOpen()} onOpenChange={setDialogOpen}>
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
