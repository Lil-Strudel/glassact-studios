import { Show, createSignal } from "solid-js";
import { useMutation, useQueryClient } from "@tanstack/solid-query";
import {
  Button,
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
  FileUpload,
  showToast,
} from "@glassact/ui";
import { IoCloudUploadOutline, IoDownloadOutline } from "solid-icons/io";
import type { ProjectStatus, SandblastFileFormat } from "@glassact/data";
import { PERMISSION_ACTIONS } from "@glassact/data";
import {
  getSandblastDownloadUrl,
  postSandblastFileOpts,
} from "../../queries/inlay";
import { postUploadOpts } from "../../queries/upload";
import { isApiError } from "../../utils/is-api-error";
import { useUserContext } from "../../providers/user";

interface SandblastFileCardProps {
  inlayUuid: string;
  inlayName: string;
  sandblastFileUrl: string | null;
  // The owning dealership's preferred delivery format, shown to whoever
  // uploads. Guidance only — no format is rejected.
  sandblastFileFormat?: SandblastFileFormat;
  projectStatus: ProjectStatus;
  // Query keys to refresh after an upload, so both the project grid and the
  // inlay page stay current wherever this is rendered from.
  onUploaded: () => void;
  layout?: "stacked" | "card";
}

export function SandblastFileCard(props: SandblastFileCardProps) {
  const userContext = useUserContext();
  const queryClient = useQueryClient();

  const uploadMutation = useMutation(() => postUploadOpts());
  const sandblastMutation = useMutation(() => postSandblastFileOpts());

  const [isDownloading, setIsDownloading] = createSignal(false);
  const [dialogOpen, setDialogOpen] = createSignal(false);

  const hasSandblast = () => !!props.sandblastFileUrl;
  const preferredFormat = () => props.sandblastFileFormat?.toUpperCase();
  const canUpload = () =>
    props.projectStatus !== "draft" &&
    userContext.can(PERMISSION_ACTIONS.MANAGE_KANBAN);

  async function handleDownload() {
    setIsDownloading(true);
    try {
      const { url } = await getSandblastDownloadUrl(props.inlayUuid);
      window.location.href = url;
    } catch (error) {
      showToast({
        title: "Failed to download sandblast file",
        description:
          error instanceof Error && isApiError(error)
            ? (error.data?.error ?? "Unknown error")
            : "Unknown error",
        variant: "error",
      });
    } finally {
      setIsDownloading(false);
    }
  }

  function handleUploaded(url: string | null | string[]) {
    const finalUrl = Array.isArray(url) ? url[0] : url;
    if (!finalUrl) return;
    sandblastMutation.mutate(
      { uuid: props.inlayUuid, body: { sandblast_file_url: finalUrl } },
      {
        onSuccess() {
          queryClient.invalidateQueries({ queryKey: ["inlay", props.inlayUuid] });
          props.onUploaded();
          showToast({
            title: "Sandblast file uploaded",
            description: `Attached to ${props.inlayName}.`,
            variant: "success",
          });
          setDialogOpen(false);
        },
        onError(error) {
          showToast({
            title: "Failed to attach sandblast file",
            description: isApiError(error)
              ? (error?.data?.error ?? "Unknown error")
              : "Unknown error",
            variant: "error",
          });
        },
      },
    );
  }

  return (
    <Show when={hasSandblast() || canUpload()}>
      <div
        class={
          props.layout === "card"
            ? "space-y-3 rounded-lg border p-4"
            : "flex flex-col gap-2"
        }
      >
        <Show when={props.layout === "card"}>
          <h2 class="text-sm font-semibold text-gray-900">Sandblast file</h2>
        </Show>

        <Show when={hasSandblast()}>
          <Button
            variant="outline"
            size="sm"
            class="w-full"
            disabled={isDownloading()}
            onClick={handleDownload}
          >
            <IoDownloadOutline size={16} class="mr-1" />
            {isDownloading() ? "Preparing..." : "Download Sandblast File"}
          </Button>
        </Show>

        <Show when={canUpload()}>
          <Dialog open={dialogOpen()} onOpenChange={setDialogOpen}>
            <DialogTrigger as={Button} variant="ghost" size="sm" class="w-full">
              <IoCloudUploadOutline size={16} class="mr-1" />
              {hasSandblast()
                ? "Replace Sandblast File"
                : "Upload Sandblast File"}
            </DialogTrigger>
            <DialogContent>
              <DialogHeader>
                <DialogTitle>
                  {hasSandblast() ? "Replace" : "Upload"} Sandblast File
                </DialogTitle>
              </DialogHeader>
              <p class="text-sm text-gray-600">
                Upload the sandblasting file for{" "}
                <span class="font-semibold">{props.inlayName}</span>. The
                dealership will be able to download it from this project.
              </p>
              <Show when={preferredFormat()}>
                {(format) => (
                  <p class="text-sm text-gray-600">
                    This dealership wants sandblasting files as{" "}
                    <span class="font-semibold">{format()}</span>.
                  </p>
                )}
              </Show>
              <div class="mt-4">
                <FileUpload
                  uploadPath="sandblast"
                  accept=".svg,.dxf,.ai,.eps,.pdf,.png,.jpg,.jpeg"
                  uploadFn={uploadMutation.mutateAsync}
                  onUrlChange={handleUploaded}
                />
              </div>
            </DialogContent>
          </Dialog>

          <Show when={preferredFormat()}>
            {(format) => (
              <p class="text-center text-xs text-gray-500">
                {format()} preferred
              </p>
            )}
          </Show>
        </Show>
      </div>
    </Show>
  );
}
