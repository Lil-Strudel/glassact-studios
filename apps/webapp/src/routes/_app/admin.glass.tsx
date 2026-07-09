import { createFileRoute } from "@tanstack/solid-router";
import { For, Show, createMemo, createSignal } from "solid-js";
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
import { GlassColor, GET } from "@glassact/data";
import { IoTrashOutline } from "solid-icons/io";
import { createForm } from "@tanstack/solid-form";
import { z } from "zod";
import { useMutation, useQuery, useQueryClient } from "@tanstack/solid-query";
import {
  getGlassColorsAdminOpts,
  postGlassColorOpts,
  patchGlassColorOpts,
  deleteGlassColorOpts,
} from "../../queries/glass-colors";
import { zodStringNumber } from "../../utils/zod-string-number";

export const Route = createFileRoute("/_app/admin/glass")({
  component: RouteComponent,
});

const HEX_REGEX = /^#[0-9a-fA-F]{6}$/;

const formSchema = z.object({
  name: z.string().min(1).max(255),
  hex: z.string().regex(HEX_REGEX, "Must be a hex color like #aabbcc"),
  family: z.string().max(255).optional().or(z.literal("")),
  sort_order: z
    .string()
    .min(1)
    .refine(...zodStringNumber),
  is_active: z.boolean(),
});

type FormSchema = z.infer<typeof formSchema>;

interface GlassColorFormProps {
  defaultValues?: FormSchema;
  onSubmit: (data: FormSchema) => void;
  isLoading: boolean;
  submitButtonLabel: string;
}

function GlassColorForm(props: GlassColorFormProps) {
  const form = createForm(() => ({
    defaultValues: props.defaultValues ?? {
      name: "",
      hex: "#000000",
      family: "",
      sort_order: "0",
      is_active: true,
    },
    validators: {
      onBlur: formSchema,
    },
    onSubmit: async ({ value }) => {
      props.onSubmit(value);
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
          <Form.TextField field={field} label="Name" placeholder="Color name" />
        )}
      />

      <form.Field
        name="hex"
        children={(field) => (
          <div class="flex flex-col gap-2">
            <Form.TextField field={field} label="Hex" placeholder="#aabbcc" />
            <input
              type="color"
              value={
                HEX_REGEX.test(field().state.value)
                  ? field().state.value
                  : "#000000"
              }
              onInput={(e) => field().handleChange(e.currentTarget.value)}
              class="h-8 w-16 cursor-pointer rounded border border-gray-300"
              aria-label="Pick a color"
            />
          </div>
        )}
      />

      <form.Field
        name="family"
        children={(field) => (
          <Form.TextField
            field={field}
            label="Family (optional)"
            placeholder="e.g. Blues"
          />
        )}
      />

      <form.Field
        name="sort_order"
        children={(field) => (
          <Form.TextField field={field} label="Sort Order" />
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
        {props.isLoading
          ? `${props.submitButtonLabel}...`
          : props.submitButtonLabel}
      </Button>
    </form>
  );
}

const defaultColumns: ColumnDef<GET<GlassColor>>[] = [
  {
    id: "actions",
    enableHiding: false,
    header: "Actions",
    cell: (props) => {
      const queryClient = useQueryClient();
      const deleteMutation = useMutation(() =>
        deleteGlassColorOpts(props.row.original.uuid),
      );

      return (
        <div class="flex gap-2">
          <EditButton item={props.row.original} />
          <Button
            variant="ghost"
            size="icon"
            onClick={() => {
              deleteMutation.mutate(undefined, {
                onSuccess: () => {
                  queryClient.invalidateQueries({ queryKey: ["glass-colors"] });
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
    accessorKey: "hex",
    header: "Color",
    cell: (info) => {
      const hex = info.getValue() as string;
      return (
        <div class="flex items-center gap-2">
          <span
            class="inline-block h-5 w-5 rounded border border-gray-300"
            style={{ "background-color": hex }}
          />
          <span class="font-mono text-sm">{hex}</span>
        </div>
      );
    },
  },
  {
    accessorKey: "family",
    header: "Family",
    cell: (info) => (info.getValue() as string | null) ?? "—",
  },
  {
    accessorKey: "sort_order",
    header: "Sort",
    cell: (info) => info.getValue(),
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
  item: GET<GlassColor>;
}

function EditButton(props: EditButtonProps) {
  const [isOpen, setIsOpen] = createSignal(false);
  const queryClient = useQueryClient();
  const patchMutation = useMutation(() => patchGlassColorOpts(props.item.uuid));

  const handleSubmit = (data: FormSchema) => {
    patchMutation.mutate(
      {
        name: data.name,
        hex: data.hex,
        family: (data.family as string) || null,
        sort_order: Number(data.sort_order),
        is_active: data.is_active,
      },
      {
        onSuccess: () => {
          queryClient.invalidateQueries({ queryKey: ["glass-colors"] });
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
          <DialogTitle>Edit Glass Color</DialogTitle>
        </DialogHeader>
        <GlassColorForm
          defaultValues={{
            name: props.item.name,
            hex: props.item.hex,
            family: props.item.family ?? "",
            sort_order: String(props.item.sort_order),
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
  const query = useQuery(() => getGlassColorsAdminOpts({ limit: 99, offset: 0 }));
  const postMutation = useMutation(() => postGlassColorOpts());
  const queryClient = useQueryClient();
  const [isCreateDialogOpen, setIsCreateDialogOpen] = createSignal(false);

  const table = createMemo(() =>
    createSolidTable({
      get data() {
        return query.data?.items ?? [];
      },
      columns: defaultColumns,
      getCoreRowModel: getCoreRowModel(),
      getPaginationRowModel: getPaginationRowModel(),
      getFilteredRowModel: getFilteredRowModel(),
      state: {
        globalFilter: filterValue(),
      },
      onGlobalFilterChange: setFilterValue,
    }),
  );

  const handleCreateSubmit = (data: FormSchema) => {
    postMutation.mutate(
      {
        name: data.name,
        hex: data.hex,
        family: (data.family as string) || null,
        sort_order: Number(data.sort_order),
        is_active: data.is_active,
      },
      {
        onSuccess: () => {
          queryClient.invalidateQueries({ queryKey: ["glass-colors"] });
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
          <DialogTrigger as={Button}>Add a new glass color</DialogTrigger>
          <DialogContent>
            <DialogHeader>
              <DialogTitle>Add a new glass color</DialogTitle>
            </DialogHeader>
            <GlassColorForm
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
            <For each={table().getHeaderGroups()}>
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
              when={table().getRowModel().rows?.length}
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
              <For each={table().getRowModel().rows}>
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
          onClick={() => table().previousPage()}
          disabled={!table().getCanPreviousPage()}
        >
          Previous
        </Button>
        <Button
          variant="outline"
          size="sm"
          onClick={() => table().nextPage()}
          disabled={!table().getCanNextPage()}
        >
          Next
        </Button>
      </div>
    </div>
  );
}
