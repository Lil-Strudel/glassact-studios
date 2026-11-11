import { getCatalogSandblastUrl } from "../../queries/catalog-browse";
import { PresignedDownloadButton } from "../presigned-download-button";

interface DownloadSandblastButtonProps {
  catalogItemUuid: string;
  class?: string;
}

export function DownloadSandblastButton(props: DownloadSandblastButtonProps) {
  return (
    <PresignedDownloadButton
      fetchUrl={() => getCatalogSandblastUrl(props.catalogItemUuid)}
      label="Download sandblast file"
      errorTitle="Failed to download sandblast file"
      class={props.class}
    />
  );
}
