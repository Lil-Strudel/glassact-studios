import { createFileRoute } from "@tanstack/solid-router";
import { For, Show, createMemo, createSignal } from "solid-js";
import {
  Badge,
  Button,
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

export const Route = createFileRoute("/_app/admin/support")({
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
        triggerClass="inline-flex h-8 items-center rounded-md border border-input bg-background px-3 text-xs font-medium hover:bg-accent"
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
        class="text-destructive hover:text-destructive"
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
    cell: (info) => info.getValue(),
  },
  {
    accessorKey: "category",
    header: "Category",
    cell: (info) => info.getValue(),
  },
  {
    accessorKey: "sort_order",
    header: "Sort",
    cell: (info) => info.getValue(),
  },
  {
    accessorKey: "is_published",
    header: "Status",
    cell: (info) => {
      const isPublished = info.getValue() as boolean;
      return (
        <Badge variant={isPublished ? "default" : "secondary"}>
          {isPublished ? "Published" : "Draft"}
        </Badge>
      );
    },
  },
];

function RouteComponent() {
  const [filterValue, setFilterValue] = createSignal("");
  const query = useQuery(() => getSupportArticlesAdminOpts());

  const table = createMemo(() =>
    createSolidTable({
      get data() {
        return query.data ?? [];
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

  return (
    <div>
      <div class="mb-6">
        <h1 class="text-2xl font-bold">Support Articles</h1>
        <p class="text-gray-600">
          Manage the knowledge base. Drafts are listed here but hidden from the
          support page.
        </p>
      </div>

      <div class="flex items-center justify-between py-4">
        <TextFieldRoot value={filterValue()} onChange={setFilterValue}>
          <TextField placeholder="Filter by title..." class="max-w-sm" />
        </TextFieldRoot>

        <ArticleFormDialog triggerClass="inline-flex h-10 items-center rounded-md bg-primary px-4 text-sm font-medium text-primary-foreground hover:bg-primary/90">
          Add a new article
        </ArticleFormDialog>
      </div>

      <div class="rounded-md border">
        <Table>
          <TableHeader>
            <For each={table().getHeaderGroups()}>
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
