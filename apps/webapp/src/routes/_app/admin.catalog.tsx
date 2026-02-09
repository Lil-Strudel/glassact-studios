import { createFileRoute, Link } from "@tanstack/solid-router";
import { For, Show, createSignal } from "solid-js";
import {
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
import { CatalogItem, GET } from "@glassact/data";
import { IoPencilOutline, IoTrashOutline } from "solid-icons/io";
import { useMutation, useQuery, useQueryClient } from "@tanstack/solid-query";
import {
  getCatalogListOpts,
  deleteCatalogOpts,
} from "../../queries/catalog";

export const Route = createFileRoute("/_app/admin/catalog")({
  component: RouteComponent,
});

const defaultColumns: ColumnDef<GET<CatalogItem>>[] = [
  {
    id: "actions",
    enableHiding: false,
    header: "Actions",
    cell: (props) => {
      const queryClient = useQueryClient();
      const deleteMutation = useMutation(() => deleteCatalogOpts(props.row.original.uuid));

      return (
        <div class="flex gap-2">
          <Button
            variant="ghost"
            size="icon"
            as={Link}
            to={`/admin/catalog/${props.row.original.uuid}`}
          >
            <IoPencilOutline size={20} />
          </Button>
          <Button
            variant="ghost"
            size="icon"
            onClick={async () => {
              if (
                window.confirm(
                  "Are you sure you want to delete this catalog item?"
                )
              ) {
                deleteMutation.mutate(undefined, {
                  onSuccess: () => {
                    queryClient.invalidateQueries({ queryKey: ["catalog"] });
                  },
                });
              }
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
    accessorKey: "catalog_code",
    header: "Code",
    cell: (info) => (
      <code class="text-xs font-mono bg-gray-100 px-2 py-1 rounded">
        {info.getValue() as string}
      </code>
    ),
  },
  {
    accessorKey: "name",
    header: "Name",
    cell: (info) => info.getValue() as string,
  },
  {
    accessorKey: "category",
    header: "Category",
    cell: (info) => info.getValue(),
  },
  {
    accessorFn: (row) =>
      `${row.default_width}x${row.default_height} (${row.min_width}-${row.default_width} x ${row.min_height}-${row.default_height})`,
    id: "dimensions",
    header: "Dimensions",
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

function RouteComponent() {
  const [filterValue, setFilterValue] = createSignal("");
  const [showInactive, setShowInactive] = createSignal(false);

  const query = useQuery(
    getCatalogListOpts({
      search: filterValue(),
      isActive: !showInactive(),
      limit: 50,
      offset: 0,
    })
  );

  const table = createSolidTable({
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
  });

  return (
    <div>
      <div class="flex items-center justify-between py-4 gap-4">
        <div class="flex items-center gap-4">
          <TextFieldRoot
            value={filterValue()}
            onChange={(value) => setFilterValue(value)}
          >
            <TextField
              placeholder="Search by code or name..."
              class="max-w-sm"
            />
          </TextFieldRoot>

          <label class="flex items-center gap-2 text-sm">
            <input
              type="checkbox"
              checked={showInactive()}
              onChange={(e) => setShowInactive(e.currentTarget.checked)}
              class="rounded border-gray-300"
            />
            Include inactive
          </label>
        </div>

        <Button as={Link} to="/admin/catalog/create">
          Create new catalog item
        </Button>
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
                    {query.isLoading ? "Loading..." : "No results."}
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
