import { createSignal } from "solid-js";
import {
  Button,
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  showToast,
} from "@glassact/ui";
import { useMutation, useQueryClient } from "@tanstack/solid-query";
import type { GET, SupportArticle } from "@glassact/data";
import { deleteSupportArticleOpts } from "../../queries/support";
import { isApiError } from "../../utils/is-api-error";
import { ArticleFormDialog } from "./article-form-dialog";

interface ArticleAdminControlsProps {
  article: GET<SupportArticle>;
}

export function ArticleAdminControls(props: ArticleAdminControlsProps) {
  const queryClient = useQueryClient();
  const [confirmOpen, setConfirmOpen] = createSignal(false);
  const deleteArticle = useMutation(() => deleteSupportArticleOpts());

  const handleDelete = () => {
    deleteArticle.mutate(props.article.uuid, {
      onSuccess() {
        queryClient.invalidateQueries({ queryKey: ["support"] });
        showToast({
          title: "Article deleted",
          description: "The support content has been removed.",
          variant: "success",
        });
        setConfirmOpen(false);
      },
      onError(error) {
        if (isApiError(error)) {
          showToast({
            title: "Failed to delete article",
            description: error?.data?.error ?? "Unknown error",
            variant: "error",
          });
        }
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
        type="button"
        variant="outline"
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
            undone.
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
