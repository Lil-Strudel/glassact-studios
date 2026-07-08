import { Show } from "solid-js";
import { Button, Form, showToast } from "@glassact/ui";
import { createForm } from "@tanstack/solid-form";
import { z } from "zod";
import { useMutation, useQueryClient } from "@tanstack/solid-query";
import { postInlayUpdateOpts } from "../../queries/manufacturing";
import { isApiError } from "../../utils/is-api-error";

const AddUpdateSchema = z.object({
  update_type: z.enum(["info", "issue"]),
  message: z.string().min(1, "Message is required"),
});

interface AddInlayUpdateFormProps {
  inlayUuid: string;
  onSuccess?: () => void;
  onCancel?: () => void;
}

export function AddInlayUpdateForm(props: AddInlayUpdateFormProps) {
  const queryClient = useQueryClient();
  const addUpdate = useMutation(() => postInlayUpdateOpts());

  const form = createForm(() => ({
    defaultValues: {
      update_type: "issue" as "info" | "issue",
      message: "",
    } as z.output<typeof AddUpdateSchema>,
    validators: {
      onSubmit: AddUpdateSchema,
    },
    onSubmit: async ({ value }) => {
      addUpdate.mutate(
        {
          inlayUuid: props.inlayUuid,
          body: {
            update_type: value.update_type,
            message: value.message,
          },
        },
        {
          onSuccess() {
            form.reset();
            queryClient.invalidateQueries({
              queryKey: ["inlay", props.inlayUuid, "updates"],
            });
            showToast({
              title: "Update posted",
              description: "The update has been added to the timeline.",
              variant: "success",
            });
            props.onSuccess?.();
          },
          onError(error) {
            if (isApiError(error)) {
              showToast({
                title: "Failed to post update",
                description: error?.data?.error ?? "Unknown error",
                variant: "error",
              });
            }
          },
        },
      );
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
        name="update_type"
        children={(field) => (
          <Form.Select
            field={field}
            label="Type"
            options={[
              { value: "issue", label: "Issue" },
              { value: "info", label: "Info" },
            ]}
            placeholder="Select type"
          />
        )}
      />

      <form.Field
        name="message"
        children={(field) => (
          <Form.TextArea
            field={field}
            label="Message"
            placeholder="Describe what happened..."
          />
        )}
      />

      <div class="flex gap-2">
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
        <Button type="submit" class="flex-1" disabled={addUpdate.isPending}>
          {addUpdate.isPending ? "Posting..." : "Post Update"}
        </Button>
      </div>
    </form>
  );
}
