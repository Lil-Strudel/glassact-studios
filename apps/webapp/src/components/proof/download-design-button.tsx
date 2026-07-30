import { Show, createSignal } from "solid-js";
import { Button, showToast } from "@glassact/ui";
import { IoDownloadOutline } from "solid-icons/io";
import { getProofDesignDownloadUrl } from "../../queries/proof";
import { isApiError } from "../../utils/is-api-error";

interface DownloadDesignButtonProps {
  proofUuid: string;
}

/**
 * Downloads a proof's design asset via a presigned URL.
 *
 * Linking straight at the asset with `<a download>` does not work: browsers
 * ignore the attribute cross-origin, so the SVG opens in a tab instead of
 * saving. The API signs a URL carrying a Content-Disposition instead.
 */
export function DownloadDesignButton(props: DownloadDesignButtonProps) {
  const [isDownloading, setIsDownloading] = createSignal(false);

  async function handleDownload() {
    setIsDownloading(true);
    try {
      const { url } = await getProofDesignDownloadUrl(props.proofUuid);
      window.location.href = url;
    } catch (error) {
      showToast({
        title: "Failed to download design",
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
      onClick={handleDownload}
      disabled={isDownloading()}
    >
      <IoDownloadOutline class="mr-2" size={16} />
      <Show when={isDownloading()} fallback="Download Design">
        Preparing...
      </Show>
    </Button>
  );
}
