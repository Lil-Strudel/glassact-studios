import { createFileRoute, Link } from "@tanstack/solid-router";
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
import { Dealership, GET } from "@glassact/data";
import { IoBuildOutline } from "solid-icons/io";
import { createForm } from "@tanstack/solid-form";
import { z } from "zod";
import { useMutation, useQuery, useQueryClient } from "@tanstack/solid-query";
import {
  getDealershipsOpts,
  postDealershipOpts,
} from "../../queries/dealership";
import { isApiError } from "../../utils/is-api-error";

export const Route = createFileRoute("/_app/admin/dealerships")({
  component: RouteComponent,
});

const defaultColumns: ColumnDef<GET<Dealership>>[] = [
  {
    id: "actions",
    enableHiding: false,
    header: "Edit",
    cell: (props) => {
      return (
        <Button
          variant="ghost"
          size="icon"
          as={Link}
          to={`/dealership/${props.row.original.uuid}`}
        >
          <IoBuildOutline size={24} />
        </Button>
      );
    },
  },
  {
    accessorKey: "name",
    header: "Name",
    cell: (info) => info.getValue(),
    footer: (info) => info.column.id,
  },
  {
    accessorKey: "address",
    header: "Address",
    cell: ({
      row: {
        original: { address },
      },
    }) =>
      `${address.street}, ${address.city}, ${address.state} ${address.postal_code}`,
    footer: (info) => info.column.id,
  },
];

const formSchema = z.object({
  name: z.string().min(1),
  requires_payment_before_shipping: z.boolean(),
  address: z.object({
    street: z.string().min(1),
    street_ext: z.string(),
    city: z.string().min(1),
    state: z.string().min(1),
    postal_code: z.string().min(1),
    country: z.string().min(1),
    latitude: z.preprocess(Number, z.number().gt(-90).lt(90)),
    longitude: z.preprocess(Number, z.number().gt(-180).lt(180)),
  }),
});

function RouteComponent() {
  const query = useQuery(() => getDealershipsOpts());
  const queryClient = useQueryClient();

  const postDealership = useMutation(postDealershipOpts);

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

  const handleDialogOpenChange = (isOpen: boolean) => {
    setDialogOpen(isOpen);
    if (!isOpen) {
      setTimeout(() => form.reset(), 300);
    }
  };

  const form = createForm(() => ({
    defaultValues: {
      name: "",
      requires_payment_before_shipping: false,
      address: {
        street: "",
        street_ext: "",
        city: "",
        state: "",
        postal_code: "",
        country: "",
        latitude: "" as unknown as number,
        longitude: "" as unknown as number,
      },
    },
    validators: {
      onSubmit: formSchema,
    },
    onSubmit: async ({ value }) => {
      const output = formSchema.parse(value);
      postDealership.mutate(output, {
        onSuccess() {
          setDialogOpen(false);
          showToast({
            title: "Created new dealership!",
            description: `${value.name} was created.`,
            variant: "success",
          });
        },
        onError(error) {
          if (isApiError(error)) {
            showToast({
              title: "Problem creating new dealership...",
              description: error?.data?.error ?? "Unknown error",
              variant: "error",
            });
          }
        },
        onSettled() {
          queryClient.invalidateQueries({ queryKey: ["dealership"] });
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
        <Dialog open={dialogOpen()} onOpenChange={handleDialogOpenChange}>
          <DialogTrigger>
            <Button>Add a new dealership</Button>
          </DialogTrigger>
          <DialogContent>
            <DialogHeader>
              <DialogTitle>Add a new dealership</DialogTitle>
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
                name="requires_payment_before_shipping"
                children={(field) => (
                  <Form.Checkbox
                    field={field}
                    label="Require payment before shipping"
                    description="Projects for this dealership show a notice that they will not ship until the invoice is paid. This does not block shipping."
                  />
                )}
              />

              <div class="border-t pt-4">
                <h3 class="text-sm font-medium text-gray-900 mb-3">Address</h3>

                <Form.AddressField
                  form={form}
                  name="address"
                  apiKey={import.meta.env.VITE_GOOGLE_MAPS_API_KEY}
                  label="Search address"
                />
              </div>

              <Button type="submit" disabled={postDealership.isPending}>
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
