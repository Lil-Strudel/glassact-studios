import { createSignal, JSX } from "solid-js";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@glassact/ui";
import type { GET, SupportArticle } from "@glassact/data";
import { ArticleForm } from "./article-form";

interface ArticleFormDialogProps {
  article?: GET<SupportArticle>;
  triggerClass?: string;
  children: JSX.Element;
}

export function ArticleFormDialog(props: ArticleFormDialogProps) {
  const [open, setOpen] = createSignal(false);

  return (
    <Dialog open={open()} onOpenChange={setOpen}>
      <DialogTrigger class={props.triggerClass}>{props.children}</DialogTrigger>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>
            {props.article ? "Edit Article" : "New Support Article"}
          </DialogTitle>
        </DialogHeader>
        <ArticleForm
          article={props.article}
          onSuccess={() => setOpen(false)}
          onCancel={() => setOpen(false)}
        />
      </DialogContent>
    </Dialog>
  );
}
