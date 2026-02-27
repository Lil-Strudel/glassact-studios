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
  FileUpload,
  showToast,
} from "@glassact/ui";
import { useMutation, useQueryClient } from "@tanstack/solid-query";
import { postProofOpts } from "../../queries/proof";
import { isApiError } from "../../utils/is-api-error";

interface CreateProofDialogProps {
  inlayUuid: string;
  onProofCreated?: () => void;
}

const inputClass =
  "h-9 w-full rounded-md border border-input bg-transparent px-3 py-1 text-sm shadow-xs transition-shadow focus-visible:outline-none focus-visible:ring-[1.5px] focus-visible:ring-ring";

export default function CreateProofDialog(props: CreateProofDialogProps) {
  const queryClient = useQueryClient();
  const createProof = useMutation(() => postProofOpts());

  const [isOpen, setIsOpen] = createSignal(false);
  const [designAssetUrl, setDesignAssetUrl] = createSignal("");
  const [width, setWidth] = createSignal<number | undefined>();
  const [height, setHeight] = createSignal<number | undefined>();
  const [priceGroupId, setPriceGroupId] = createSignal<number | undefined>();

  function resetForm() {
    setDesignAssetUrl("");
    setWidth(undefined);
    setHeight(undefined);
    setPriceGroupId(undefined);
  }

  function handleSubmit(e: SubmitEvent) {
    e.preventDefault();

    const url = designAssetUrl();
    const w = width();
    const h = height();

    if (!url) {
      showToast({
        title: "Validation error",
        description: "Please upload a design file.",
        variant: "error",
      });
      return;
    }

    if (!w || w <= 0 || !h || h <= 0) {
      showToast({
        title: "Validation error",
        description: "Width and height must be greater than 0.",
        variant: "error",
      });
      return;
    }

    const pgId = priceGroupId();

    createProof.mutate(
      {
        inlayUuid: props.inlayUuid,
        body: {
          design_asset_url: url,
          width: w,
          height: h,
          ...(pgId != null ? { price_group_id: pgId } : {}),
        },
      },
      {
        onSuccess() {
          setIsOpen(false);
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
          resetForm();
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
  }

  return (
    <Dialog open={isOpen()} onOpenChange={setIsOpen}>
      <DialogTrigger as={Button} variant="outline" size="sm">
        Send Proof
      </DialogTrigger>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Create Proof</DialogTitle>
        </DialogHeader>
        <form onSubmit={handleSubmit} class="flex flex-col gap-4">
          <FileUpload
            onUrlChange={(url) => {
              if (typeof url === "string") {
                setDesignAssetUrl(url);
              } else if (Array.isArray(url) && url.length > 0) {
                setDesignAssetUrl(url[0]);
              } else {
                setDesignAssetUrl("");
              }
            }}
            uploadPath="proofs"
            accept=".pdf,.png,.jpg,.jpeg,.svg"
            fileTypeLabel="PDF, Image, or SVG"
            label="Design File"
            description="Upload the proof design file"
          />

          <div class="flex flex-col gap-1.5">
            <label class="text-sm font-medium" for="proof-width">
              Width
            </label>
            <input
              id="proof-width"
              type="number"
              required
              min="0.01"
              step="any"
              class={inputClass}
              value={width() ?? ""}
              onInput={(e) => {
                const val = parseFloat(e.currentTarget.value);
                setWidth(Number.isNaN(val) ? undefined : val);
              }}
            />
          </div>

          <div class="flex flex-col gap-1.5">
            <label class="text-sm font-medium" for="proof-height">
              Height
            </label>
            <input
              id="proof-height"
              type="number"
              required
              min="0.01"
              step="any"
              class={inputClass}
              value={height() ?? ""}
              onInput={(e) => {
                const val = parseFloat(e.currentTarget.value);
                setHeight(Number.isNaN(val) ? undefined : val);
              }}
            />
          </div>

          <div class="flex flex-col gap-1.5">
            <label class="text-sm font-medium" for="proof-price-group">
              Price Group ID (optional)
            </label>
            <input
              id="proof-price-group"
              type="number"
              step="1"
              class={inputClass}
              value={priceGroupId() ?? ""}
              onInput={(e) => {
                const val = parseInt(e.currentTarget.value, 10);
                setPriceGroupId(Number.isNaN(val) ? undefined : val);
              }}
            />
          </div>

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
