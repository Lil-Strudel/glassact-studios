import { createForm } from "@tanstack/solid-form";
import { z } from "zod";
import { useMutation, useQueryClient } from "@tanstack/solid-query";
import { Button, Form, showToast } from "@glassact/ui";
import { patchInlayOpts } from "../../queries/inlay";
import { postUploadOpts } from "../../queries/upload";
import { isApiError } from "../../utils/is-api-error";

interface EditCustomInfoFormProps {
  inlayUuid: string;
  description: string;
  imageUrls: string[];
  onDone: () => void;
}

const EditCustomInfoSchema = z.object({
  description: z.string().min(1, "Description is required"),
  image_urls: z.array(z.string()),
});

export default function EditCustomInfoForm(props: EditCustomInfoFormProps) {
  const queryClient = useQueryClient();
  const patchInlay = useMutation(patchInlayOpts);
  const uploadMutation = useMutation(postUploadOpts);

  const form = createForm(() => ({
    defaultValues: {
      description: props.description,
      image_urls: props.imageUrls,
    },
    validators: {
      onSubmit: EditCustomInfoSchema,
    },
    onSubmit: async ({ value }) => {
      patchInlay.mutate(
        {
          uuid: props.inlayUuid,
          body: {
            description: value.description,
            image_urls: value.image_urls,
          },
        },
        {
          onSuccess() {
            queryClient.invalidateQueries({
              queryKey: ["inlay", props.inlayUuid],
            });
            showToast({
              title: "Inlay updated",
              description: "Your changes have been saved.",
              variant: "success",
            });
            props.onDone();
          },
          onError(error) {
            showToast({
              title: "Failed to update inlay",
              description: isApiError(error)
                ? (error.data?.error ?? "Unknown error")
                : "An unexpected error occurred.",
              variant: "error",
            });
          },
        },
      );
    },
  }));

  return (
    <form
      onSubmit={(e) => {
        e.preventDefault();
        e.stopPropagation();
        form.handleSubmit();
      }}
      class="flex flex-col gap-4"
    >
      <form.Field
        name="description"
        children={(field) => (
          <Form.TextArea
            field={field}
            label="Description"
            placeholder="Describe the desired design in detail..."
          />
        )}
      />
      <form.Field
        name="image_urls"
        children={(field) => (
          <Form.FileUpload
            field={field}
            label="Reference pictures"
            multiple
            uploadPath="inlay-references"
            accept=".png,.jpg,.jpeg,.gif,.webp"
            fileTypeLabel="image"
            description="Add or remove pictures showing the desired design."
            uploadFn={uploadMutation.mutateAsync}
          />
        )}
      />
      <div class="flex justify-end gap-2">
        <Button
          type="button"
          variant="outline"
          size="sm"
          onClick={() => props.onDone()}
          disabled={patchInlay.isPending}
        >
          Cancel
        </Button>
        <Button type="submit" size="sm" disabled={patchInlay.isPending}>
          {patchInlay.isPending ? "Saving..." : "Save"}
        </Button>
      </div>
    </form>
  );
}
