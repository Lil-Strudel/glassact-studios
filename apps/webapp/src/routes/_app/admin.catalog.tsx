import {
  createFileRoute,
  Link,
  stripSearchParams,
} from "@tanstack/solid-router";
import { For, Show, createMemo } from "solid-js";
import { z } from "zod";
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
import { useClampPage } from "../../hooks/use-clamp-page";
import { useDebouncedSearchInput } from "../../hooks/use-debounced-search-input";
import { useScrollToTopOnChange } from "../../hooks/use-scroll-to-top-on-change";
import {
  adminCatalogSearchDefaults,
  adminCatalogSearchSchema,
  type AdminCatalogSearchInput,
} from "../../utils/list-search";
import { PageHeader } from "@glassact/ui";
import { EmptyState } from "@glassact/ui";

const PAGE_SIZE = 50;

function toCatalogListParams(
  search: z.output<typeof adminCatalogSearchSchema>,
) {
  return {
    search: search.q,
    // Omitting the filter returns both; the server treats a present
    // is_active as strict equality, so passing false would hide the
    // active items instead of adding the inactive ones.
    isActive: search.inactive ? undefined : true,
    limit: PAGE_SIZE,
    offset: (search.page - 1) * PAGE_SIZE,
  };
}

export const Route = createFileRoute("/_app/admin/catalog")({
  validateSearch: adminCatalogSearchSchema,
  search: { middlewares: [stripSearchParams(adminCatalogSearchDefaults)] },
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
            <IoPencilOutline size={18} />
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
            <IoTrashOutline size={18} />
          </Button>
        </div>
      );
    },
  },
  {
    accessorKey: "catalog_code",
    header: "Code",
    cell: (info) => (
      <code class="whitespace-nowrap rounded border border-border bg-secondary px-2 py-0.5 font-mono text-xs tabular-nums text-muted-foreground">
        {info.getValue() as string}
      </code>
    ),
  },
  {
    accessorKey: "name",
    header: "Name",
    cell: (info) => (
      <span class="font-medium text-foreground">
        {info.getValue() as string}
      </span>
    ),
  },
  {
    accessorKey: "category",
    header: "Category",
    cell: (info) => (
      <span class="capitalize text-muted-foreground">
        {info.getValue() as string}
      </span>
    ),
  },
  {
    accessorFn: (row) =>
      `${row.default_width}x${row.default_height} (${row.min_width}-${row.default_width} x ${row.min_height}-${row.default_height})`,
    id: "dimensions",
    header: "Dimensions",
    cell: (info) => (
      <span class="whitespace-nowrap tabular-nums text-muted-foreground">
        {info.getValue() as string}
      </span>
    ),
  },
  {
    accessorKey: "display_order",
    header: "Rank",
    cell: (info) => {
      const rank = info.getValue() as number | null;
      return (
        <Show
          when={rank != null}
          fallback={<span class="text-xs text-faint">Unranked</span>}
        >
          <Badge variant="gilt" class="tabular-nums">
            #{rank}
          </Badge>
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
        <Badge variant={isActive ? "success" : "outline"} lozenge>
          {isActive ? "Yes" : "No"}
        </Badge>
      );
    },
  },
];

function RouteComponent() {
  const search = Route.useSearch();
  const navigate = Route.useNavigate();

  function updateSearch(patch: AdminCatalogSearchInput) {
    navigate({
      search: (prev) => ({ ...prev, ...patch }),
      replace: true,
      resetScroll: false,
    });
  }

  const [searchText, setSearchText] = useDebouncedSearchInput(
    () => search().q,
    (q) => updateSearch({ q, page: 1 }),
  );

  const query = useQuery(() =>
    getCatalogListOpts(toCatalogListParams(search())),
  );

  const totalPages = createMemo(() =>
    Math.max(1, Math.ceil((query.data?.total ?? 0) / PAGE_SIZE)),
  );

  useClampPage({
    page: () => search().page,
    pageCount: totalPages,
    hasSettled: () => query.isSuccess && !query.isPlaceholderData,
    setPage: (page) => updateSearch({ page }),
  });

  let tableTop: HTMLDivElement | undefined;
  useScrollToTopOnChange(
    () => search().page,
    () => tableTop,
  );

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
    <div class="space-y-6">
      <PageHeader
        eyebrow="Administration"
        title="Catalog"
        actions={
          <>
            <Button variant="outline" as={Link} to="/admin/catalog/order">
              Edit best sellers
            </Button>
            <Button as={Link} to="/admin/catalog/create">
              Create new catalog item
            </Button>
          </>
        }
      />
      <div class="flex flex-wrap items-center gap-4">
        <TextFieldRoot value={searchText()} onChange={setSearchText}>
          <TextField placeholder="Search by code or name..." class="max-w-sm" />
        </TextFieldRoot>

        <label class="flex items-center gap-2 text-sm text-muted-foreground">
          <input
            type="checkbox"
            checked={search().inactive}
            onChange={(e) =>
              updateSearch({ inactive: e.currentTarget.checked, page: 1 })
            }
            class="size-4 rounded border-input accent-primary"
          />
          Include inactive
        </label>
      </div>

      <div
        ref={tableTop}
        class="scroll-mt-6 overflow-hidden rounded-lg border border-border bg-card"
      >
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
                <TableRow class="hover:bg-transparent">
                  <TableCell colSpan={defaultColumns.length} class="p-0">
                    <Show
                      when={!query.isLoading}
                      fallback={
                        <p class="py-12 text-center text-sm text-faint">
                          Loading...
                        </p>
                      }
                    >
                      <EmptyState
                        title="No catalog items found"
                        description="Try a different code or name, or create the first catalog item."
                      />
                    </Show>
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

      <div class="flex items-center justify-end gap-4">
        <span class="text-sm tabular-nums text-faint">
          Page {search().page} of {totalPages()} ({query.data?.total ?? 0}{" "}
          items)
        </span>
        <div class="flex items-center gap-2">
          <Button
            variant="outline"
            size="sm"
            onClick={() => updateSearch({ page: search().page - 1 })}
            disabled={search().page <= 1}
          >
            Previous
          </Button>
          <Button
            variant="outline"
            size="sm"
            onClick={() => updateSearch({ page: search().page + 1 })}
            disabled={search().page >= totalPages()}
          >
            Next
          </Button>
        </div>
      </div>
    </div>
  );
}
