import { Show } from "solid-js";
import { Button, Form, showToast } from "@glassact/ui";
import { createForm } from "@tanstack/solid-form";
import { z } from "zod";
import { useMutation, useQueryClient } from "@tanstack/solid-query";
import type { GET, SupportArticle, SupportCategory } from "@glassact/data";
import {
  patchSupportArticleOpts,
  postSupportArticleOpts,
} from "../../queries/support";
import { isApiError } from "../../utils/is-api-error";

const CATEGORY_OPTIONS: { value: SupportCategory; label: string }[] = [
  { value: "installation", label: "Installation" },
  { value: "ordering", label: "Placing Orders" },
  { value: "pricing", label: "Pricing" },
  { value: "contact", label: "Contact" },
  { value: "general", label: "General" },
];

const ArticleSchema = z.object({
  category: z.enum([
    "installation",
    "ordering",
    "pricing",
    "contact",
    "general",
  ]),
  title: z.string().min(1, "Title is required"),
  body: z.string(),
  youtube_url: z
    .string()
    .trim()
    .url("Must be a valid URL")
    .or(z.literal(""))
    .optional(),
  sort_order: z.number().int(),
  is_published: z.boolean(),
});

type ArticleFormValues = z.output<typeof ArticleSchema>;

interface ArticleFormProps {
  article?: GET<SupportArticle>;
  onSuccess?: () => void;
  onCancel?: () => void;
}

export function ArticleForm(props: ArticleFormProps) {
  const queryClient = useQueryClient();
  const createArticle = useMutation(() => postSupportArticleOpts());
  const updateArticle = useMutation(() => patchSupportArticleOpts());

  const isEditing = () => props.article !== undefined;
  const isPending = () => createArticle.isPending || updateArticle.isPending;

  const form = createForm(() => ({
    defaultValues: {
      category: (props.article?.category ?? "general") as SupportCategory,
      title: props.article?.title ?? "",
      body: props.article?.body ?? "",
      youtube_url: props.article?.youtube_url ?? "",
      sort_order: props.article?.sort_order ?? 0,
      is_published: props.article?.is_published ?? true,
    } as ArticleFormValues,
    validators: {
      onSubmit: ArticleSchema,
    },
    onSubmit: async ({ value }) => {
      const body = {
        category: value.category,
        title: value.title,
        body: value.body,
        youtube_url: value.youtube_url ? value.youtube_url.trim() : null,
        sort_order: value.sort_order,
        is_published: value.is_published,
      };

      const handlers = {
        onSuccess() {
          queryClient.invalidateQueries({ queryKey: ["support"] });
          showToast({
            title: isEditing() ? "Article updated" : "Article created",
            description: "The support content has been saved.",
            variant: "success",
          });
          props.onSuccess?.();
        },
        onError(error: Error) {
          if (isApiError(error)) {
            showToast({
              title: "Failed to save article",
              description: error?.data?.error ?? "Unknown error",
              variant: "error",
            });
          }
        },
      };

      const existing = props.article;
      if (existing) {
        updateArticle.mutate({ uuid: existing.uuid, body }, handlers);
      } else {
        createArticle.mutate(body, handlers);
      }
    },
  }));

  return (
    <form
      onSubmit={(e) => {
        e.preventDefault();
        form.handleSubmit();
      }}
      class="space-y-3"
    >
      <form.Field
        name="category"
        children={(field) => (
          <Form.Select
            field={field}
            label="Category"
            options={CATEGORY_OPTIONS}
            placeholder="Select category"
          />
        )}
      />

      <form.Field
        name="title"
        children={(field) => (
          <Form.TextField field={field} label="Title" placeholder="Article title" />
        )}
      />

      <form.Field
        name="youtube_url"
        children={(field) => (
          <Form.TextField
            field={field}
            label="YouTube URL (optional)"
            placeholder="https://www.youtube.com/watch?v=..."
            description="Paste a YouTube link to embed a video with this article."
          />
        )}
      />

      <form.Field
        name="body"
        children={(field) => (
          <Form.TextArea
            field={field}
            label="Content"
            placeholder="Write the tip or how-to here..."
            description="Markdown is supported (headings, bold, links, lists)."
          />
        )}
      />

      <div class="flex gap-3">
        <form.Field
          name="sort_order"
          children={(field) => (
            <Form.NumberField
              field={field}
              label="Sort order"
              int
              class="w-32"
              fullWidth={false}
            />
          )}
        />
        <form.Field
          name="is_published"
          children={(field) => (
            <Form.Checkbox
              field={field}
              label="Published"
              class="pt-6"
            />
          )}
        />
      </div>

      <div class="flex gap-2 pt-2">
        <Show when={props.onCancel}>
          <Button
            type="button"
            variant="outline"
            class="flex-1"
            onClick={() => {
              form.reset();
              props.onCancel?.();
            }}
          >
            Cancel
          </Button>
        </Show>
        <Button type="submit" class="flex-1" disabled={isPending()}>
          {isPending()
            ? "Saving..."
            : isEditing()
              ? "Save Changes"
              : "Create Article"}
        </Button>
      </div>
    </form>
  );
}
