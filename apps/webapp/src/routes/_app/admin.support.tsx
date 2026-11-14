import { createFileRoute, stripSearchParams } from "@tanstack/solid-router";
import { For, Show, createSignal } from "solid-js";
import {
  Badge,
  Button,
  buttonVariants,
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
  TextField,
  TextFieldRoot,
  showToast,
} from "@glassact/ui";
import {
  ColumnDef,
  createSolidTable,
  flexRender,
  getCoreRowModel,
  getFilteredRowModel,
  getPaginationRowModel,
} from "@tanstack/solid-table";
import { useMutation, useQuery, useQueryClient } from "@tanstack/solid-query";
import type { GET, SupportArticle } from "@glassact/data";
import {
  deleteSupportArticleOpts,
  getSupportArticlesAdminOpts,
  patchSupportArticleOpts,
} from "../../queries/support";
import { ArticleFormDialog } from "../../components/support/article-form-dialog";
import { isApiError } from "../../utils/is-api-error";
import { useClampPage } from "../../hooks/use-clamp-page";
import { useDebouncedSearchInput } from "../../hooks/use-debounced-search-input";
import { useScrollToTopOnChange } from "../../hooks/use-scroll-to-top-on-change";
import {
  listSearchDefaults,
  listSearchSchema,
  type ListSearchInput,
} from "../../utils/list-search";
import {
  onTablePaginationChange,
  tablePaginationState,
} from "../../utils/table-pagination";
import { PageHeader } from "@glassact/ui";
import { EmptyState } from "@glassact/ui";

export const Route = createFileRoute("/_app/admin/support")({
  validateSearch: listSearchSchema,
  search: { middlewares: [stripSearchParams(listSearchDefaults)] },
  component: RouteComponent,
});

function toErrorMessage(error: unknown) {
  if (error instanceof Error && isApiError(error)) {
    return error.data?.error ?? "Unknown error";
  }
  return "Unknown error";
}

interface ArticleActionsProps {
  article: GET<SupportArticle>;
}

function ArticleActions(props: ArticleActionsProps) {
  const queryClient = useQueryClient();
  const [confirmOpen, setConfirmOpen] = createSignal(false);
  const patchArticle = useMutation(() => patchSupportArticleOpts());
  const deleteArticle = useMutation(() => deleteSupportArticleOpts());

  // The public support list is published-only, so both lists have to refresh.
  const invalidate = () =>
    queryClient.invalidateQueries({ queryKey: ["support"] });

  const togglePublished = () => {
    const nextPublished = !props.article.is_published;

    patchArticle.mutate(
      {
        uuid: props.article.uuid,
        body: { is_published: nextPublished },
      },
      {
        onSuccess() {
          invalidate();
          showToast({
            title: nextPublished ? "Article published" : "Article unpublished",
            description: nextPublished
              ? `"${props.article.title}" is now live on the support page.`
              : `"${props.article.title}" is hidden from the support page.`,
            variant: "success",
          });
        },
        onError(error) {
          showToast({
            title: "Problem updating article...",
            description: toErrorMessage(error),
            variant: "error",
          });
        },
      },
    );
  };

  const handleDelete = () => {
    deleteArticle.mutate(props.article.uuid, {
      onSuccess() {
        invalidate();
        showToast({
          title: "Article deleted",
          description: "The support content has been removed.",
          variant: "success",
        });
        setConfirmOpen(false);
      },
      onError(error) {
        showToast({
          title: "Failed to delete article",
          description: toErrorMessage(error),
          variant: "error",
        });
      },
    });
  };

  return (
    <div class="flex items-center gap-2">
      <ArticleFormDialog
        article={props.article}
        triggerClass={buttonVariants({ variant: "outline", size: "sm" })}
      >
        Edit
      </ArticleFormDialog>

      <Button
        variant="ghost"
        size="sm"
        onClick={togglePublished}
        disabled={patchArticle.isPending}
      >
        {props.article.is_published ? "Unpublish" : "Publish"}
      </Button>

      <Button
        variant="ghost"
        size="sm"
        class="text-error-foreground hover:bg-error hover:text-error-foreground"
        onClick={() => setConfirmOpen(true)}
      >
        Delete
      </Button>

      <Dialog open={confirmOpen()} onOpenChange={setConfirmOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Delete this article?</DialogTitle>
          </DialogHeader>
          <p class="text-sm text-muted-foreground">
            "{props.article.title}" will be permanently removed. This cannot be
            undone — unpublish it instead if you only want to hide it.
          </p>
          <DialogFooter class="gap-2">
            <Button
              type="button"
              variant="outline"
              onClick={() => setConfirmOpen(false)}
            >
              Cancel
            </Button>
            <Button
              type="button"
              variant="destructive"
              disabled={deleteArticle.isPending}
              onClick={handleDelete}
            >
              {deleteArticle.isPending ? "Deleting..." : "Delete"}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}

const defaultColumns: ColumnDef<GET<SupportArticle>>[] = [
  {
    id: "actions",
    enableHiding: false,
    header: "Actions",
    cell: (info) => <ArticleActions article={info.row.original} />,
  },
  {
    accessorKey: "title",
    header: "Title",
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
      <span class="text-muted-foreground">{info.getValue() as string}</span>
    ),
  },
  {
    accessorKey: "sort_order",
    header: "Sort",
    cell: (info) => (
      <span class="tabular-nums text-muted-foreground">
        {info.getValue() as number}
      </span>
    ),
  },
  {
    accessorKey: "is_published",
    header: "Status",
    cell: (info) => {
      const isPublished = info.getValue() as boolean;
      return (
        <Badge variant={isPublished ? "success" : "draft"} lozenge>
          {isPublished ? "Published" : "Draft"}
        </Badge>
      );
    },
  },
];

function RouteComponent() {
  const search = Route.useSearch();
  const navigate = Route.useNavigate();

  function updateSearch(patch: ListSearchInput) {
    navigate({
      search: (prev) => ({ ...prev, ...patch }),
      replace: true,
      resetScroll: false,
    });
  }

  const [filterText, setFilterText] = useDebouncedSearchInput(
    () => search().q,
    (q) => updateSearch({ q, page: 1 }),
  );

  const query = useQuery(() => getSupportArticlesAdminOpts());

  const table = createSolidTable({
    get data() {
      return query.data ?? [];
    },
    columns: defaultColumns,
    getCoreRowModel: getCoreRowModel(),
    getPaginationRowModel: getPaginationRowModel(),
    getFilteredRowModel: getFilteredRowModel(),
    state: {
      get pagination() {
        return tablePaginationState(search().page);
      },
      get globalFilter() {
        return search().q;
      },
    },
    onPaginationChange: onTablePaginationChange(
      () => search().page,
      (page) => updateSearch({ page }),
    ),
    autoResetPageIndex: false,
  });

  useClampPage({
    page: () => search().page,
    pageCount: () => table.getPageCount(),
    hasSettled: () => query.isSuccess,
    setPage: (page) => updateSearch({ page }),
  });

  let tableTop: HTMLDivElement | undefined;
  useScrollToTopOnChange(
    () => search().page,
    () => tableTop,
  );

  return (
    <div class="space-y-6">
      <PageHeader
        eyebrow="Administration"
        title="Support Articles"
        description="Manage the knowledge base. Drafts are listed here but hidden from the support page."
      />

      <div class="flex flex-wrap items-center justify-between gap-3">
        <TextFieldRoot value={filterText()} onChange={setFilterText}>
          <TextField placeholder="Filter by title..." class="max-w-sm" />
        </TextFieldRoot>

        <ArticleFormDialog triggerClass={buttonVariants()}>
          Add a new article
        </ArticleFormDialog>
      </div>

      <div
        ref={tableTop}
        class="scroll-mt-6 overflow-hidden rounded-lg border border-border bg-card"
      >
        <Table>
          <TableHeader>
            <For each={table.getHeaderGroups()}>
              {(headerGroup) => (
                <TableRow>
                  <For each={headerGroup.headers}>
                    {(header) => (
                      <TableHead>
                        {header.isPlaceholder
                          ? null
                          : flexRender(
                              header.column.columnDef.header,
                              header.getContext(),
                            )}
                      </TableHead>
                    )}
                  </For>
                </TableRow>
              )}
            </For>
          </TableHeader>
          <TableBody>
            <Show
              when={table.getRowModel().rows?.length}
              fallback={
                <TableRow class="hover:bg-transparent">
                  <TableCell colSpan={defaultColumns.length} class="p-0">
                    <EmptyState
                      title="No articles found"
                      description="Try a different title, or add the first article."
                    />
                  </TableCell>
                </TableRow>
              }
            >
              <For each={table.getRowModel().rows}>
                {(row) => (
                  <TableRow>
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

      <div class="flex items-center justify-end gap-2">
        <span class="mr-2 text-sm tabular-nums text-faint">
          Page {table.getState().pagination.pageIndex + 1} of{" "}
          {Math.max(table.getPageCount(), 1)}
        </span>
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
