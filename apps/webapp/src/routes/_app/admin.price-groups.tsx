import { createFileRoute } from "@tanstack/solid-router";
import { For, Show, createSignal } from "solid-js";
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
  Badge,
} from "@glassact/ui";
import {
  createSolidTable,
  flexRender,
  getCoreRowModel,
  ColumnDef,
  getPaginationRowModel,
  getFilteredRowModel,
} from "@tanstack/solid-table";
import { PriceGroup, GET } from "@glassact/data";
import { IoTrashOutline } from "solid-icons/io";
import { createForm } from "@tanstack/solid-form";
import { z } from "zod";
import { useMutation, useQuery, useQueryClient } from "@tanstack/solid-query";
import {
  getPriceGroupsOpts,
  postPriceGroupOpts,
  patchPriceGroupOpts,
  deletePriceGroupOpts,
} from "../../queries/price-group";

export const Route = createFileRoute("/_app/admin/price-groups")({
  component: RouteComponent,
});

const formSchema = z.object({
  name: z.string().min(1).max(255),
  base_price_cents: z.number().positive().int(),
  description: z.string().max(1000).optional().or(z.literal("")),
  is_active: z.boolean(),
});

type FormSchema = z.infer<typeof formSchema>;

const defaultColumns: ColumnDef<GET<PriceGroup>>[] = [
  {
    id: "actions",
    enableHiding: false,
    header: "Actions",
    cell: (props) => {
      const queryClient = useQueryClient();
      const deleteMutation = useMutation(() => deletePriceGroupOpts(props.row.original.uuid));

      return (
        <div class="flex gap-2">
          <EditButton item={props.row.original} />
          <Button
            variant="ghost"
            size="icon"
            onClick={async () => {
              deleteMutation.mutate(undefined, {
                onSuccess: () => {
                  queryClient.invalidateQueries({ queryKey: ["price-groups"] });
                },
              });
            }}
            disabled={deleteMutation.isPending}
          >
            <IoTrashOutline size={20} />
          </Button>
        </div>
      );
    },
  },
  {
    accessorKey: "name",
    header: "Name",
    cell: (info) => info.getValue(),
  },
  {
    accessorKey: "base_price_cents",
    header: "Base Price",
    cell: (info) => {
      const cents = info.getValue() as number;
      return `$${(cents / 100).toFixed(2)}`;
    },
  },
  {
    accessorKey: "description",
    header: "Description",
    cell: (info) => {
      const desc = info.getValue() as string | null;
      if (!desc) return "—";
      return desc.length > 50 ? `${desc.substring(0, 50)}...` : desc;
    },
  },
  {
    accessorKey: "is_active",
    header: "Active",
    cell: (info) => {
      const isActive = info.getValue() as boolean;
      return (
        <Badge variant={isActive ? "default" : "secondary"}>
          {isActive ? "Yes" : "No"}
        </Badge>
      );
    },
  },
];

interface EditButtonProps {
  item: GET<PriceGroup>;
}

function EditButton(props: EditButtonProps) {
  const [isOpen, setIsOpen] = createSignal(false);
  const queryClient = useQueryClient();
  const patchMutation = useMutation(() => patchPriceGroupOpts(props.item.uuid));

  const form = createForm(() => ({
    defaultValues: {
      name: props.item.name,
      base_price_cents: props.item.base_price_cents,
      description: props.item.description ?? "",
      is_active: props.item.is_active,
    },
    onSubmit: async ({ value }) => {
      patchMutation.mutate(
        {
          name: value.name,
          base_price_cents: value.base_price_cents as number,
          description: (value.description as string) || null,
          is_active: value.is_active,
        },
        {
          onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: ["price-groups"] });
            setIsOpen(false);
            form.reset();
          },
        },
      );
    },
  }));

  return (
    <Dialog open={isOpen()} onOpenChange={setIsOpen}>
      <DialogTrigger as={Button} variant="outline" size="sm">
        Edit
      </DialogTrigger>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Edit Price Group</DialogTitle>
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
              <div class="flex flex-col gap-2">
                <label class="text-sm font-medium text-gray-900">Name</label>
                <input
                  type="text"
                  value={field().state.value}
                  onInput={(e) => field().handleChange(e.currentTarget.value)}
                  class="rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2"
                />
              </div>
            )}
          />

              <form.Field
                name="base_price_cents"
                children={(field) => (
                  <div class="flex flex-col gap-2">
                    <label class="text-sm font-medium text-gray-900">Base Price (cents)</label>
                    <input
                      type="number"
                      value={field().state.value}
                      onInput={(e) => field().handleChange(Number(e.currentTarget.value))}
                      class="rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2"
                    />
                  </div>
                )}
              />

          <form.Field
            name="description"
            children={(field) => (
              <div class="flex flex-col gap-2">
                <label class="text-sm font-medium text-gray-900">Description (optional)</label>
                <textarea
                  value={field().state.value}
                  onInput={(e) => field().handleChange(e.currentTarget.value)}
                  class="rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2"
                  rows={3}
                />
              </div>
            )}
          />

          <form.Field
            name="is_active"
            children={(field) => (
              <label class="flex items-center gap-2">
                <input
                  type="checkbox"
                  checked={field().state.value}
                  onChange={(e) => field().handleChange(e.currentTarget.checked)}
                  class="rounded border-gray-300"
                />
                <span class="text-sm font-medium text-gray-900">Active</span>
              </label>
            )}
          />

          <Button type="submit" disabled={patchMutation.isPending}>
            {patchMutation.isPending ? "Updating..." : "Update"}
          </Button>
        </form>
      </DialogContent>
    </Dialog>
  );
}

function RouteComponent() {
  const [filterValue, setFilterValue] = createSignal("");
  const query = useQuery(getPriceGroupsOpts as any);
  const postMutation = useMutation(() => postPriceGroupOpts());
  const queryClient = useQueryClient();

  const table = createSolidTable({
    get data() {
      return (query.data as any)?.items ?? [];
    },
    columns: defaultColumns,
    getCoreRowModel: getCoreRowModel(),
    getPaginationRowModel: getPaginationRowModel(),
    getFilteredRowModel: getFilteredRowModel(),
    state: {
      globalFilter: filterValue(),
    },
    onGlobalFilterChange: setFilterValue,
  });

  const form = createForm(() => ({
    defaultValues: {
      name: "",
      base_price_cents: 0,
      description: "",
      is_active: true,
    },
    onSubmit: async ({ value }) => {
      postMutation.mutate(
        {
          name: value.name,
          base_price_cents: value.base_price_cents as number,
          description: (value.description as string) || null,
          is_active: value.is_active,
        },
        {
          onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: ["price-groups"] });
            form.reset();
          },
        },
      );
    },
  }));

  return (
    <div>
      <div class="flex items-center justify-between py-4">
        <TextFieldRoot
          value={filterValue()}
          onChange={(value) => setFilterValue(value)}
        >
          <TextField placeholder="Filter by name..." class="max-w-sm" />
        </TextFieldRoot>
        <Dialog>
          <DialogTrigger as={Button}>Add a new price group</DialogTrigger>
          <DialogContent>
            <DialogHeader>
              <DialogTitle>Add a new price group</DialogTitle>
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
                  <div class="flex flex-col gap-2">
                    <label class="text-sm font-medium text-gray-900">Name</label>
                    <input
                      type="text"
                      value={field().state.value}
                      onInput={(e) => field().handleChange(e.currentTarget.value)}
                      class="rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2"
                    />
                  </div>
                )}
              />

              <form.Field
                name="base_price_cents"
                children={(field) => (
                  <div class="flex flex-col gap-2">
                    <label class="text-sm font-medium text-gray-900">Base Price (cents)</label>
                    <input
                      type="number"
                      value={field().state.value}
                      onInput={(e) => field().handleChange(Number(e.currentTarget.value))}
                      class="rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2"
                    />
                  </div>
                )}
              />

                <form.Field
                  name="description"
                  children={(field) => (
                    <div class="flex flex-col gap-2">
                      <label class="text-sm font-medium text-gray-900">Description (optional)</label>
                      <textarea
                        value={field().state.value}
                        onInput={(e) => field().handleChange(e.currentTarget.value)}
                        class="rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2"
                        rows={3}
                      />
                    </div>
                  )}
                />

              <form.Field
                name="is_active"
                children={(field) => (
                  <label class="flex items-center gap-2">
                    <input
                      type="checkbox"
                      checked={field().state.value}
                      onChange={(e) => field().handleChange(e.currentTarget.checked)}
                      class="rounded border-gray-300"
                    />
                    <span class="text-sm font-medium text-gray-900">Active</span>
                  </label>
                )}
              />

              <Button type="submit" disabled={postMutation.isPending}>
                {postMutation.isPending ? "Adding..." : "Add"}
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
