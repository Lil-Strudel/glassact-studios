import { Show, createSignal } from "solid-js";
import { Button, showToast } from "@glassact/ui";
import { IoDownloadOutline } from "solid-icons/io";
import { isApiError } from "../utils/is-api-error";

interface PresignedDownloadButtonProps {
  fetchUrl: () => Promise<{ url: string }>;
  label: string;
  errorTitle: string;
  class?: string;
}

/**
 * Linking straight at an S3 asset with `<a download>` does not work: browsers
 * ignore the attribute cross-origin, so the SVG opens in a tab instead of
 * saving. The API signs a URL carrying a Content-Disposition instead.
 */
export function PresignedDownloadButton(props: PresignedDownloadButtonProps) {
  const [isDownloading, setIsDownloading] = createSignal(false);

  async function handleDownload() {
    setIsDownloading(true);
    try {
      const { url } = await props.fetchUrl();
      window.location.href = url;
    } catch (error) {
      showToast({
        title: props.errorTitle,
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

  return (
    <Button
      variant="outline"
      size="sm"
      class={props.class}
      onClick={handleDownload}
      disabled={isDownloading()}
    >
      <IoDownloadOutline class="mr-2" size={16} />
      <Show when={isDownloading()} fallback={props.label}>
        Preparing...
      </Show>
    </Button>
  );
}
