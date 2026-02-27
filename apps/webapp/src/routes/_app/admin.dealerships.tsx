import { createFileRoute, Link } from "@tanstack/solid-router";
import { For, Show } from "solid-js";
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
import { useMutation, useQuery } from "@tanstack/solid-query";
import {
  getDealershipsOpts,
  postDealershipOpts,
} from "../../queries/dealership";

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

  const form = createForm(() => ({
    defaultValues: {
      name: "",
      address: {
        street: "",
        street_ext: "",
        city: "",
        state: "",
        postal_code: "",
        country: "",
        latitude: undefined as unknown as Number,
        longitude: undefined as unknown as Number,
      },
    },
    validators: {
      onSubmit: formSchema,
    },
    onSubmit: async ({ value }) => {
      const output = formSchema.parse(value);
      postDealership.mutate(output, { onSuccess() {} });
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
        <Dialog>
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

              <div class="border-t pt-4">
                <h3 class="text-sm font-medium text-gray-900 mb-3">Address</h3>

                <div class="grid grid-cols-1 gap-4">
                  <form.Field
                    name="address.street"
                    children={(field) => (
                      <Form.TextField field={field} label="Street" />
                    )}
                  />

                  <form.Field
                    name="address.street_ext"
                    children={(field) => (
                      <Form.TextField field={field} label="Street Extension" />
                    )}
                  />

                  <div class="grid grid-cols-2 gap-4">
                    <form.Field
                      name="address.city"
                      children={(field) => (
                        <Form.TextField field={field} label="City" />
                      )}
                    />

                    <form.Field
                      name="address.state"
                      children={(field) => (
                        <Form.TextField field={field} label="State" />
                      )}
                    />
                  </div>

                  <div class="grid grid-cols-2 gap-4">
                    <form.Field
                      name="address.postal_code"
                      children={(field) => (
                        <Form.TextField field={field} label="Postal Code" />
                      )}
                    />

                    <form.Field
                      name="address.country"
                      children={(field) => (
                        <Form.TextField field={field} label="Country" />
                      )}
                    />
                  </div>

                  <div class="grid grid-cols-2 gap-4">
                    <form.Field
                      name="address.latitude"
                      children={(field) => (
                        <Form.TextField field={field} label="Latitude" />
                      )}
                    />

                    <form.Field
                      name="address.longitude"
                      children={(field) => (
                        <Form.TextField field={field} label="Longitude" />
                      )}
                    />
                  </div>
                </div>
              </div>

              <Button type="submit">Add</Button>
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
