import { createFileRoute, Link } from "@tanstack/solid-router";
import { For, Show, createMemo, createSignal } from "solid-js";
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
} from "@tanstack/solid-table";
import { CatalogItem, GET } from "@glassact/data";
import { IoPencilOutline, IoTrashOutline } from "solid-icons/io";
import { useMutation, useQuery, useQueryClient } from "@tanstack/solid-query";
import { getCatalogListOpts, deleteCatalogOpts } from "../../queries/catalog";
import { useDebounce } from "../../hooks/use-debounce";

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
      const deleteCatalog = useMutation(() => deleteCatalogOpts());

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
            onClick={() => {
              if (
                window.confirm(
                  "Are you sure you want to deactivate this catalog item?",
                )
              ) {
                deleteCatalog.mutate(props.row.original.uuid, {
                  onSuccess: () => {
                    queryClient.invalidateQueries({ queryKey: ["catalog"] });
                  },
                });
              }
            }}
            disabled={deleteCatalog.isPending}
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
    accessorKey: "display_order",
    header: "Rank",
    cell: (info) => {
      const rank = info.getValue() as number | null;
      return (
        <Show
          when={rank != null}
          fallback={<span class="text-xs text-gray-400">Unranked</span>}
        >
          <Badge variant="default">#{rank}</Badge>
        </Show>
      );
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

const PAGE_SIZE = 50;

function RouteComponent() {
  const [filterValue, setFilterValue] = createSignal("");
  const [showInactive, setShowInactive] = createSignal(false);
  const [page, setPage] = createSignal(1);

  const debouncedFilterValue = useDebounce(filterValue, 300);

  const query = useQuery(() =>
    getCatalogListOpts({
      search: debouncedFilterValue(),
      // Omitting the filter returns both; the server treats a present
      // is_active as strict equality, so passing false would hide the
      // active items instead of adding the inactive ones.
      isActive: showInactive() ? undefined : true,
      limit: PAGE_SIZE,
      offset: (page() - 1) * PAGE_SIZE,
    }),
  );

  const totalPages = createMemo(() =>
    Math.max(1, Math.ceil((query.data?.total ?? 0) / PAGE_SIZE)),
  );

  // Narrowing the result set can leave you past the end of it.
  function handleSearchChange(value: string) {
    setFilterValue(value);
    setPage(1);
  }

  function handleShowInactiveChange(value: boolean) {
    setShowInactive(value);
    setPage(1);
  }

  // Paging is server-side, so the table renders exactly the rows it is given.
  const table = createMemo(() =>
    createSolidTable({
      get data() {
        return query.data?.items ?? [];
      },
      columns: defaultColumns,
      getCoreRowModel: getCoreRowModel(),
    }),
  );

  return (
    <div>
      <div class="flex items-center justify-between py-4 gap-4">
        <div class="flex items-center gap-4">
          <TextFieldRoot value={filterValue()} onChange={handleSearchChange}>
            <TextField
              placeholder="Search by code or name..."
              class="max-w-sm"
            />
          </TextFieldRoot>

          <label class="flex items-center gap-2 text-sm">
            <input
              type="checkbox"
              checked={showInactive()}
              onChange={(e) => handleShowInactiveChange(e.currentTarget.checked)}
              class="rounded border-gray-300"
            />
            Include inactive
          </label>
        </div>

        <div class="flex items-center gap-2">
          <Button variant="outline" as={Link} to="/admin/catalog/order">
            Edit best sellers
          </Button>
          <Button as={Link} to="/admin/catalog/create">
            Create new catalog item
          </Button>
        </div>
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
                    {query.isLoading ? "Loading..." : "No results."}
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

      <div class="flex items-center justify-end gap-4 py-4">
        <span class="text-sm text-gray-500">
          Page {page()} of {totalPages()} ({query.data?.total ?? 0} items)
        </span>
        <div class="flex items-center gap-2">
          <Button
            variant="outline"
            size="sm"
            onClick={() => setPage((current) => Math.max(1, current - 1))}
            disabled={page() <= 1}
          >
            Previous
          </Button>
          <Button
            variant="outline"
            size="sm"
            onClick={() =>
              setPage((current) => Math.min(totalPages(), current + 1))
            }
            disabled={page() >= totalPages()}
          >
            Next
          </Button>
        </div>
      </div>
    </div>
  );
}
