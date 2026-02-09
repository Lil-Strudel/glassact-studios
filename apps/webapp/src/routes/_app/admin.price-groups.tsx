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

interface PriceGroupFormProps {
  defaultValues?: FormSchema;
  onSubmit: (data: FormSchema) => void;
  isLoading: boolean;
  submitButtonLabel: string;
}

function PriceGroupForm(props: PriceGroupFormProps) {
  const form = createForm(() => ({
    defaultValues: props.defaultValues ?? {
      name: "",
      base_price_cents: 0,
      description: "",
      is_active: true,
    },
    onSubmit: async ({ value }) => {
      props.onSubmit(value as FormSchema);
    },
  }));

  return (
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
          <Form.TextField field={field} label="Name" placeholder="Price group name" />
        )}
      />

      <form.Field
        name="base_price_cents"
        children={(field) => (
          <Form.TextField
            field={field}
            label="Base Price (cents)"
            placeholder="0"
          />
        )}
      />

      <form.Field
        name="description"
        children={(field) => (
          <Form.TextArea
            field={field}
            label="Description (optional)"
            placeholder="Optional description"
          />
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

      <Button type="submit" disabled={props.isLoading}>
        {props.isLoading ? `${props.submitButtonLabel}...` : props.submitButtonLabel}
      </Button>
    </form>
  );
}

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

  const handleSubmit = (data: FormSchema) => {
    patchMutation.mutate(
      {
        name: data.name,
        base_price_cents: data.base_price_cents as number,
        description: (data.description as string) || null,
        is_active: data.is_active,
      },
      {
        onSuccess: () => {
          queryClient.invalidateQueries({ queryKey: ["price-groups"] });
          setIsOpen(false);
        },
      },
    );
  };

  return (
    <Dialog open={isOpen()} onOpenChange={setIsOpen}>
      <DialogTrigger as={Button} variant="outline" size="sm">
        Edit
      </DialogTrigger>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Edit Price Group</DialogTitle>
        </DialogHeader>
        <PriceGroupForm
          defaultValues={{
            name: props.item.name,
            base_price_cents: props.item.base_price_cents,
            description: props.item.description ?? "",
            is_active: props.item.is_active,
          }}
          onSubmit={handleSubmit}
          isLoading={patchMutation.isPending}
          submitButtonLabel="Update"
        />
      </DialogContent>
    </Dialog>
  );
}

function RouteComponent() {
  const [filterValue, setFilterValue] = createSignal("");
  const query = useQuery(getPriceGroupsOpts as any);
  const postMutation = useMutation(() => postPriceGroupOpts());
  const queryClient = useQueryClient();
  const [isCreateDialogOpen, setIsCreateDialogOpen] = createSignal(false);

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

  const handleCreateSubmit = (data: FormSchema) => {
    postMutation.mutate(
      {
        name: data.name,
        base_price_cents: data.base_price_cents as number,
        description: (data.description as string) || null,
        is_active: data.is_active,
      },
      {
        onSuccess: () => {
          queryClient.invalidateQueries({ queryKey: ["price-groups"] });
          setIsCreateDialogOpen(false);
        },
      },
    );
  };

  return (
    <div>
      <div class="flex items-center justify-between py-4">
        <TextFieldRoot
          value={filterValue()}
          onChange={(value) => setFilterValue(value)}
        >
          <TextField placeholder="Filter by name..." class="max-w-sm" />
        </TextFieldRoot>
        <Dialog open={isCreateDialogOpen()} onOpenChange={setIsCreateDialogOpen}>
          <DialogTrigger as={Button}>Add a new price group</DialogTrigger>
          <DialogContent>
            <DialogHeader>
              <DialogTitle>Add a new price group</DialogTitle>
            </DialogHeader>
            <PriceGroupForm
              onSubmit={handleCreateSubmit}
              isLoading={postMutation.isPending}
              submitButtonLabel="Add"
            />
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
