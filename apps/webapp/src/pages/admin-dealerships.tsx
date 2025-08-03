import { type Component, createSignal, For, Show } from "solid-js";
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
import { A } from "@solidjs/router";

export const defaultData: GET<Dealership>[] = Array.from(new Array(100)).map(
  (_, index) => ({
    id: index,
    uuid: "uuid" + index,
    name: `Test ${index + 1}`,
    address: "646 W 80 N, Orem, UT 84057",
    location: [12, 12],
    created_at: "",
    updated_at: "",
    version: 1,
  }),
);

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
          as="a"
          href={`/dealership/${props.row.original.uuid}`}
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
    cell: (info) => info.getValue(),
    footer: (info) => info.column.id,
  },
];

const AdminDealerships: Component = () => {
  const [data, setData] = createSignal(defaultData);

  const table = createSolidTable({
    get data() {
      return data();
    },
    columns: defaultColumns,
    getCoreRowModel: getCoreRowModel(),
    getPaginationRowModel: getPaginationRowModel(),
    getFilteredRowModel: getFilteredRowModel(),
  });

  const form = createForm(() => ({
    defaultValues: {
      name: "",
      address: "",
    },
    validators: {
      onSubmit: z.object({
        name: z.string().min(1),
        address: z.string().min(1),
      }),
    },
    onSubmit: async ({ value }) => {
      console.log(value);
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
              <form.Field
                name="address"
                children={(field) => (
                  <Form.TextField field={field} label="Address" />
                )}
              />
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
};

export default AdminDealerships;
