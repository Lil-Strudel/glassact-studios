import { createSignal, Show } from "solid-js";
import {
  Dialog,
  DialogTrigger,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogFooter,
  DialogClose,
  Button,
  Form,
  showToast,
} from "@glassact/ui";
import { useMutation, useQueryClient } from "@tanstack/solid-query";
import { createForm } from "@tanstack/solid-form";
import { z } from "zod";
import { postProofOpts } from "../../queries/proof";
import { postUploadOpts } from "../../queries/upload";
import { isApiError } from "../../utils/is-api-error";
import PriceGroupCombobox from "../price-group-combobox";

interface CreateProofDialogProps {
  inlayUuid: string;
  onProofCreated?: () => void;
}

const CreateProofSchema = z.object({
  design_asset_url: z.string().min(1, "Design file is required"),
  width: z.number().positive("Width must be greater than 0"),
  height: z.number().positive("Height must be greater than 0"),
  price_group_id: z.number().int().optional(),
});

export default function CreateProofDialog(props: CreateProofDialogProps) {
  const queryClient = useQueryClient();
  const createProof = useMutation(() => postProofOpts());
  const uploadMutation = useMutation(() => postUploadOpts());

  const [isOpen, setIsOpen] = createSignal(false);

  const form = createForm(() => ({
    defaultValues: {
      design_asset_url: "",
      width: "" as unknown as number,
      height: "" as unknown as number,
      price_group_id: undefined,
    } as z.output<typeof CreateProofSchema>,
    validators: {
      onSubmit: CreateProofSchema,
    },
    onSubmit: async ({ value }) => {
      createProof.mutate(
        {
          inlayUuid: props.inlayUuid,
          body: {
            design_asset_url: value.design_asset_url,
            width: value.width!,
            height: value.height!,
            ...(value.price_group_id
              ? { price_group_id: value.price_group_id }
              : {}),
          },
        },
        {
          onSuccess() {
            setIsOpen(false);
            form.reset();
            queryClient.invalidateQueries({
              queryKey: ["inlay", props.inlayUuid, "proofs"],
            });
            queryClient.invalidateQueries({
              queryKey: ["inlay", props.inlayUuid, "chats"],
            });
            props.onProofCreated?.();
            showToast({
              title: "Proof created",
              description: "The proof has been sent successfully.",
              variant: "success",
            });
          },
          onError(error) {
            if (isApiError(error)) {
              showToast({
                title: "Failed to create proof",
                description: error.data?.error ?? "Unknown error",
                variant: "error",
              });
            } else {
              showToast({
                title: "Failed to create proof",
                description: "An unexpected error occurred.",
                variant: "error",
              });
            }
          },
        },
      );
    },
  }));

  return (
    <Dialog open={isOpen()} onOpenChange={setIsOpen}>
      <DialogTrigger as={Button} variant="outline" size="sm">
        Send Proof
      </DialogTrigger>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Create Proof</DialogTitle>
        </DialogHeader>
        <form
          onSubmit={(e) => {
            e.preventDefault();
            form.handleSubmit();
          }}
          class="flex flex-col gap-4"
        >
          <form.Field
            name="design_asset_url"
            children={(field) => (
              <Form.FileUpload
                field={field}
                label="Design File"
                uploadPath="proofs"
                accept=".pdf,.png,.jpg,.jpeg,.svg"
                fileTypeLabel="PDF, Image, or SVG"
                description="Upload the proof design file"
                multiple={false}
                uploadFn={uploadMutation.mutateAsync}
              />
            )}
          />

          <form.Field
            name="width"
            children={(field) => (
              <Form.NumberField
                field={field}
                label="Width"
                placeholder="e.g., 100.50"
                decimalPlaces={2}
              />
            )}
          />

          <form.Field
            name="height"
            children={(field) => (
              <Form.NumberField
                field={field}
                label="Height"
                placeholder="e.g., 50.00"
                decimalPlaces={2}
              />
            )}
          />

          <form.Field
            name="price_group_id"
            children={(field) => (
              <div class="flex flex-col gap-2">
                <PriceGroupCombobox field={field} />
                <p class="text-xs text-muted-foreground">
                  Optional. Leave blank to use default pricing, or select to
                  override.
                </p>
              </div>
            )}
          />

          <DialogFooter class="flex justify-end gap-3">
            <DialogClose as={Button} variant="outline">
              Cancel
            </DialogClose>
            <Button type="submit" disabled={createProof.isPending}>
              <Show when={createProof.isPending} fallback="Send Proof">
                Sending...
              </Show>
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}
